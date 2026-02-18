package engine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"testDB/internal/types"
)

type IndexMeta struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"` // "hash" | "btree"
	Fields     []string `json:"fields"`
	Unique     bool     `json:"unique"`
	Status     string   `json:"status"` // "building" | "ready"
	CreatedAt  int64    `json:"createdAt"`
	UpdatedAt  int64    `json:"updatedAt"`
}

type HashIndex struct {
	Meta    IndexMeta
	Entries map[string][]string // key -> docIDs
}

type BTreeIndex struct {
	Meta IndexMeta

	// only for single field btree
	Kind string // "number" | "time"
	Keys []float64
	Map  map[float64][]string
}

func indexName(indexType string, fields []string) string {
	return indexType + ":" + strings.Join(fields, ",")
}

func (c *Collection) indexesPath(cfg Config) string {
	// data/databases/<db>/collections/<coll>/indexes.json
	collDir := filepath.Dir(c.DataFile) // .../<coll>
	return filepath.Join(collDir, "indexes.json")
}

func (c *Collection) ensureIndexMaps() {
	if c.IndexesHash == nil {
		c.IndexesHash = map[string]*HashIndex{}
	}
	if c.IndexesBTree == nil {
		c.IndexesBTree = map[string]*BTreeIndex{}
	}
	if c.IndexMetas == nil {
		c.IndexMetas = map[string]IndexMeta{}
	}
}

func (c *Collection) loadIndexMetas(cfg Config) {
	c.ensureIndexMaps()
	p := c.indexesPath(cfg)
	b, err := os.ReadFile(p)
	if err != nil || len(b) == 0 {
		return
	}
	var metas []IndexMeta
	if json.Unmarshal(b, &metas) != nil {
		return
	}
	for _, m := range metas {
		c.IndexMetas[m.Name] = m
	}
}

func (c *Collection) saveIndexMetas(cfg Config) error {
	c.ensureIndexMaps()
	p := c.indexesPath(cfg)

	metas := make([]IndexMeta, 0, len(c.IndexMetas))
	for _, m := range c.IndexMetas {
		metas = append(metas, m)
	}
	sort.Slice(metas, func(i, j int) bool { return metas[i].Name < metas[j].Name })

	b, _ := json.MarshalIndent(metas, "", "  ")
	return atomicWriteFile(p, b, 0666)
}

func (e *Engine) CreateIndex(dbName, collName string, fields []string, indexType string, unique bool, background bool) error {
	db, err := e.getOrCreateDB(dbName)
	if err != nil { return err }
	c, err := db.getOrCreateCollection(e.cfg, collName)
	if err != nil { return err }

	fieldsNorm := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" { fieldsNorm = append(fieldsNorm, f) }
	}
	if len(fieldsNorm) == 0 {
		return errors.New("fields required")
	}

	indexType = strings.ToLower(strings.TrimSpace(indexType))
	if indexType != "hash" && indexType != "btree" {
		return errors.New("index type must be hash or btree")
	}
	if indexType == "btree" && len(fieldsNorm) != 1 {
		return errors.New("btree supports only single field for now")
	}

	name := indexName(indexType, fieldsNorm)

	c.mu.Lock()
	c.ensureIndexMaps()
	// already exists?
	if meta, ok := c.IndexMetas[name]; ok && meta.Status == "ready" {
		c.mu.Unlock()
		return nil
	}

	meta := IndexMeta{
		Name:      name,
		Type:      indexType,
		Fields:    fieldsNorm,
		Unique:    unique,
		Status:    "building",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	c.IndexMetas[name] = meta
	_ = c.saveIndexMetas(e.cfg)
	c.mu.Unlock()

	build := func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		// Build snapshot from existing docs
		// Copy docs (avoid mutation)
		snap := make([]types.Document, 0, len(c.Docs))
		for _, d := range c.Docs {
			snap = append(snap, d)
		}

		if indexType == "hash" {
			idx, err := buildHashIndex(meta, snap)
			if err != nil { return err }
			c.IndexesHash[name] = idx
		} else {
			idx, err := buildBTreeIndex(meta, snap)
			if err != nil { return err }
			c.IndexesBTree[name] = idx
		}

		// mark ready
		m := c.IndexMetas[name]
		m.Status = "ready"
		m.UpdatedAt = time.Now().Unix()
		c.IndexMetas[name] = m
		return c.saveIndexMetas(e.cfg)
	}

	if background {
		go func() { _ = build() }()
		return nil
	}
	return build()
}

// ---------- builders ----------

func buildHashIndex(meta IndexMeta, docs []types.Document) (*HashIndex, error) {
	idx := &HashIndex{Meta: meta, Entries: map[string][]string{}}

	seenUnique := map[string]string{} // key -> docID

	for _, d := range docs {
		id, _ := d["_id"].(string)

		key := compoundKey(d, meta.Fields)
		if meta.Unique {
			if prev, ok := seenUnique[key]; ok && prev != id {
				return nil, errors.New("unique index violation on key: " + key)
			}
			seenUnique[key] = id
		}
		idx.Entries[key] = append(idx.Entries[key], id)
	}
	return idx, nil
}

func buildBTreeIndex(meta IndexMeta, docs []types.Document) (*BTreeIndex, error) {
	field := meta.Fields[0]
	idx := &BTreeIndex{Meta: meta, Map: map[float64][]string{}}

	// detect kind (number/time) by scanning docs
	kind := ""
	for _, d := range docs {
		v, ok := getNestedField(d, field)
		if !ok || v == nil {
			continue
		}
		if _, ok := toNumber(v); ok {
			kind = "number"
			break
		}
		if _, ok := toTime(v); ok {
			kind = "time"
			break
		}
	}
	if kind == "" {
		return nil, errors.New("btree index supports only numeric or RFC3339 date fields")
	}
	idx.Kind = kind

	for _, d := range docs {
		id, _ := d["_id"].(string)
		v, ok := getNestedField(d, field)
		if !ok || v == nil {
			continue
		}

		var k float64
		if kind == "number" {
			n, ok := toNumber(v)
			if !ok { continue }
			k = n
		} else {
			t, ok := toTime(v)
			if !ok { continue }
			k = float64(t.UnixNano())
		}

		if meta.Unique {
			if ids, ok := idx.Map[k]; ok && len(ids) > 0 && ids[0] != id {
				return nil, errors.New("unique index violation on value")
			}
		}

		idx.Map[k] = append(idx.Map[k], id)
	}

	// sort keys
	keys := make([]float64, 0, len(idx.Map))
	for k := range idx.Map {
		keys = append(keys, k)
	}
	sort.Float64s(keys)
	idx.Keys = keys
	return idx, nil
}

func compoundKey(doc types.Document, fields []string) string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		v, _ := getNestedField(doc, f)
		parts = append(parts, toKeyString(v))
	}
	return strings.Join(parts, "|")
}

func toKeyString(v any) string {
	if v == nil { return "null" }
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		strings.TrimSpace(toString(v)),
		"|", "_"), "\n", " "), "\r", " "))
}
