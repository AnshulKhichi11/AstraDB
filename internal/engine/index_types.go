package engine

import(
	"sync"
)

type Index struct {
	Field   string
	Unique  bool
	Entries map[string][]string // key -> list of _id
	mu      sync.RWMutex
}

type IndexMetadata struct {
	Field  string `json:"field"`
	Unique bool   `json:"unique"`
}
