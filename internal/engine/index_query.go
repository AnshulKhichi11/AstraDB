package engine

import (
	"sort"

	"testDB/internal/types"
)

// findByID looks up a document by ID.
// NOTE: caller must NOT hold c.mu — this function acquires it internally.
func (c *Collection) findByID(id string) (types.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Support segments
	if c.useSegments && c.segmentMgr != nil {
		docs, err := c.segmentMgr.ReadAll()
		if err == nil {
			for _, d := range docs {
				if s, ok := d["_id"].(string); ok && s == id {
					return d, true
				}
			}
		}
		return nil, false
	}

	for _, d := range c.Docs {
		if s, ok := d["_id"].(string); ok && s == id {
			return d, true
		}
	}
	return nil, false
}

// candidateIDsByIndex returns candidate doc IDs using an index, if one applies.
// IMPORTANT: caller must already hold c.mu (read or write) — this function
// does NOT acquire c.mu to avoid a double-lock deadlock.
// (Query holds c.mu.RLock before calling this.)
func (c *Collection) candidateIDsByIndex(filter map[string]any) ([]string, bool) {
	// ── NO c.mu.RLock here ── caller already holds it

	// --- 1) Range (btree) ---
	for field, want := range filter {
		opMap, ok := want.(map[string]any)
		if !ok {
			continue
		}
		if hasRangeOp(opMap) {
			name := indexName("btree", []string{field})
			idx, ok := c.IndexesBTree[name]
			if !ok || idx.Meta.Status != "ready" {
				continue
			}
			ids := btreeRangeLookup(idx, opMap)
			return ids, true
		}
	}

	// --- 2) Hash (multi-field equality) ---
	bestName := ""
	bestFields := 0
	for name, meta := range c.IndexMetas {
		if meta.Type != "hash" || meta.Status != "ready" {
			continue
		}
		ok := true
		for _, f := range meta.Fields {
			if _, exists := filter[f]; !exists {
				ok = false
				break
			}
			if _, isOp := filter[f].(map[string]any); isOp {
				ok = false
				break
			}
		}
		if ok && len(meta.Fields) > bestFields {
			bestFields = len(meta.Fields)
			bestName = name
		}
	}

	if bestName != "" {
		idx := c.IndexesHash[bestName]
		if idx == nil {
			return nil, false
		}
		key := compoundKey(types.Document(filter), c.IndexMetas[bestName].Fields)
		ids := idx.Entries[key]
		return ids, true
	}

	return nil, false
}

func hasRangeOp(m map[string]any) bool {
	for k := range m {
		switch k {
		case "$gt", "$gte", "$lt", "$lte":
			return true
		}
	}
	return false
}

func btreeRangeLookup(idx *BTreeIndex, opMap map[string]any) []string {
	loSet, hiSet := false, false
	var lo, hi float64
	loInc, hiInc := false, false

	if v, ok := opMap["$gt"]; ok {
		loSet = true
		loInc = false
		lo = btreeKey(idx, v)
	}
	if v, ok := opMap["$gte"]; ok {
		loSet = true
		loInc = true
		lo = btreeKey(idx, v)
	}
	if v, ok := opMap["$lt"]; ok {
		hiSet = true
		hiInc = false
		hi = btreeKey(idx, v)
	}
	if v, ok := opMap["$lte"]; ok {
		hiSet = true
		hiInc = true
		hi = btreeKey(idx, v)
	}

	keys := idx.Keys
	start := 0
	if loSet {
		start = sort.Search(len(keys), func(i int) bool {
			if loInc {
				return keys[i] >= lo
			}
			return keys[i] > lo
		})
	}
	end := len(keys)
	if hiSet {
		end = sort.Search(len(keys), func(i int) bool {
			if hiInc {
				return keys[i] > hi
			}
			return keys[i] >= hi
		})
	}

	if start < 0 {
		start = 0
	}
	if end > len(keys) {
		end = len(keys)
	}
	if start >= end {
		return []string{}
	}

	out := []string{}
	for _, k := range keys[start:end] {
		out = append(out, idx.Map[k]...)
	}
	return out
}

func btreeKey(idx *BTreeIndex, v any) float64 {
	if idx.Kind == "number" {
		n, ok := toNumber(v)
		if ok {
			return n
		}
		return 0
	}
	t, ok := toTime(v)
	if ok {
		return float64(t.UnixNano())
	}
	return 0
}
