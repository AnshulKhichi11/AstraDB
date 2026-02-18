package types

import (
	"time"
	"fmt"
)	

type Document map[string]any

type Collection struct {
	Name     string
	DataFile string
	Docs     []Document
}

func Stringify(v any) string {
	return fmt.Sprintf("%v", v)
}

type Database struct {
	Name        string
	Collections map[string]*Collection
}

type WALEntry struct {
	TS         int64  `json:"ts"`
	Op         string `json:"op"` // insert|update|delete
	DB         string `json:"db"`
	Collection string `json:"collection"`
	Doc        any    `json:"doc,omitempty"`
	Filter     any    `json:"filter,omitempty"`
	Update     any    `json:"update,omitempty"`
	Multi      bool   `json:"multi,omitempty"`
}

type ISODate struct {
	Time time.Time
}

