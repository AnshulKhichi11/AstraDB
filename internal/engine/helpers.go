package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"testDB/internal/types"
)

func ensureDirs(cfg Config) error {
	dirs := []string{cfg.DataDir, cfg.DBsDir, cfg.WALDir}
	for _, d := range dirs {
		if err := mkdirAll(d); err != nil {
			return err
		}
	}
	if _, err := os.Stat(cfg.WALFile); os.IsNotExist(err) {
		return os.WriteFile(cfg.WALFile, []byte(""), 0666)
	}
	return nil
}

func mkdirAll(p string) error { return os.MkdirAll(p, 0755) }

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func normalizeName(s string) string {
	s = strings.TrimSpace(s)
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			out = append(out, r)
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}

func dbCollectionsDir(cfg Config, db string) string {
	return filepath.Join(cfg.DBsDir, db, "collections")
}
func collectionDir(cfg Config, db, coll string) string {
	return filepath.Join(cfg.DBsDir, db, "collections", coll)
}
func collectionDataFile(cfg Config, db, coll string) string {
	return filepath.Join(cfg.DBsDir, db, "collections", coll, "data.db")
}

func mapKeysSorted[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func applySkipLimit(in []types.Document, skip, limit int) []types.Document {
	if skip < 0 {
		skip = 0
	}
	if skip >= len(in) {
		return []types.Document{}
	}
	in = in[skip:]
	if limit > 0 && limit < len(in) {
		in = in[:limit]
	}
	return in
}

func applySort(docs []types.Document, sortSpec map[string]int) {
	if len(sortSpec) == 0 {
		return
	}
	keys := make([]string, 0, len(sortSpec))
	for k := range sortSpec {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sort.Slice(docs, func(i, j int) bool {
		for _, k := range keys {
			dir := sortSpec[k]
			ai, _ := getNestedField(docs[i], k)
			aj, _ := getNestedField(docs[j], k)
			cmp := compareAny(ai, aj)
			if cmp == 0 {
				continue
			}
			if dir >= 0 {
				return cmp < 0
			}
			return cmp > 0
		}
		return false
	})
}

func applyProjection(docs []types.Document, proj map[string]int) []types.Document {
	if len(proj) == 0 {
		return docs
	}

	// include-mode if any field == 1
	includeMode := false
	for _, v := range proj {
		if v == 1 {
			includeMode = true
			break
		}
	}

	out := make([]types.Document, 0, len(docs))
	for _, d := range docs {
		nd := types.Document{}

		if includeMode {
			for f, v := range proj {
				if v != 1 {
					continue
				}
				if val, ok := getNestedField(d, f); ok {
					setNestedField(nd, f, val)
				}
			}
			// default include _id unless explicitly excluded
			if proj["_id"] != 0 {
				if id, ok := d["_id"]; ok {
					nd["_id"] = id
				}
			}
		} else {
			// exclude-mode (fields set to 0)
			for k, v := range d {
				nd[k] = v
			}
			for f, v := range proj {
				if v == 0 {
					unsetNestedField(nd, f)
				}
			}
		}

		out = append(out, nd)
	}
	return out
}

func enforceDocSizeLimit(doc types.Document, max int) error {
	b, _ := json.Marshal(doc)
	if len(b) > max {
		return errors.New("document too large (exceeds limit)")
	}
	return nil
}

// ---------- A1: ObjectId-like sortable ID ----------
var oidCounter uint32

func NewObjectID() string {
	// 12 bytes:
	// 4 bytes unix seconds + 5 bytes random + 3 bytes counter
	b := make([]byte, 12)
	sec := uint32(time.Now().Unix())
	b[0] = byte(sec >> 24)
	b[1] = byte(sec >> 16)
	b[2] = byte(sec >> 8)
	b[3] = byte(sec)

	_, _ = rand.Read(b[4:9])

	c := atomic.AddUint32(&oidCounter, 1)
	b[9] = byte(c >> 16)
	b[10] = byte(c >> 8)
	b[11] = byte(c)

	return hex.EncodeToString(b) // 24 hex chars
}

// ---------- A1: Number strictness + JSON.Number parsing ----------
// func canonicalizeDocument(doc types.Document) types.Document {
// 	return walkCanonical(doc).(types.Document)
// }

func walkCanonical(v any) any {
	switch x := v.(type) {
	case map[string]any:
		// ISODate support: {"$date":"RFC3339"}
		if d, ok := x["$date"]; ok && len(x) == 1 {
			if s, ok := d.(string); ok {
				// store as RFC3339 string
				return s
			}
		}
		m := map[string]any{}
		for k, vv := range x {
			m[k] = walkCanonical(vv)
		}
		return types.Document(m)

	case []any:
		out := make([]any, 0, len(x))
		for _, it := range x {
			out = append(out, walkCanonical(it))
		}
		return out

	case json.Number:
		s := x.String()
		if strings.ContainsAny(s, ".eE") {
			f, err := x.Float64()
			if err == nil {
				return f
			}
			return s
		}
		i, err := x.Int64()
		if err == nil {
			return i
		}
		f, err := x.Float64()
		if err == nil {
			return f
		}
		return s

	default:
		return v
	}
}

