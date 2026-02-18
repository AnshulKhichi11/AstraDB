package engine

import (
	"sort"

	"testDB/internal/types"
)

func (c *Collection) findByID(id string) (types.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, d := range c.Docs {
		if s, ok := d["_id"].(string); ok && s == id {
			return d, true
		}
	}
	return nil, false
}

// Planner:
// 1) If range filter exists ($gt/$gte/$lt/$lte) and btree index exists -> use it
// 2) Else if equality filter matches a compound/single hash index -> use it
// 3) Else nil (fallback full scan)
func (c *Collection) candidateIDsByIndex(filter map[string]any) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// --- 1) Range (btree) ---
	for field, want := range filter {
		opMap, ok := want.(map[string]any)
		if !ok {
			continue
		}
		if hasRangeOp(opMap) {
			// find btree meta name
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
	// Find the "best" hash index: most fields matched
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
			// only direct equality (not op map) for hash usage
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
		key := compoundKey(types.Document(filter), c.IndexMetas[bestName].Fields) // cheap: uses toKeyString
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
	// bounds
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
	// binary search lower
	start := 0
	if loSet {
		start = sort.Search(len(keys), func(i int) bool {
			if loInc {
				return keys[i] >= lo
			}
			return keys[i] > lo
		})
	}
	// binary search upper
	end := len(keys)
	if hiSet {
		end = sort.Search(len(keys), func(i int) bool {
			if hiInc {
				return keys[i] > hi
			}
			return keys[i] >= hi
		})
	}

	if start < 0 { start = 0 }
	if end > len(keys) { end = len(keys) }
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
		if ok { return n }
		return 0
	}
	// time: store unixNano
	t, ok := toTime(v)
	if ok { return float64(t.UnixNano()) }
	return 0
}
