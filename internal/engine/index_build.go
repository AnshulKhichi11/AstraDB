package engine

import (
	"fmt"
)

func (idx *Index) build(c *Collection) {

	for _, doc := range c.Docs {

		val := getIndexValue(doc, idx.Field)

		if val == "" {
			continue
		}

		id := doc["_id"].(string)

		idx.Entries[val] = append(idx.Entries[val], id)
	}
}

func getIndexValue(doc map[string]any, field string) string {

	val, ok := getNestedField(doc, field)

	if !ok {
		return ""
	}

	return fmt.Sprintf("%v", val)
}
