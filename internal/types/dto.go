package types

type InsertRequest struct {
	DB         string   `json:"db"`
	Collection string   `json:"collection"`
	Data       Document `json:"data"`
}

type QueryRequest struct {
	DB         string         `json:"db"`
	Collection string         `json:"collection"`
	Filter     map[string]any `json:"filter"`
	Sort       map[string]int `json:"sort"`
	Limit      int            `json:"limit"`
	Skip       int            `json:"skip"`
	Projection map[string]int `json:"projection"` // {field:1, _id:0}
}

type UpdateRequest struct {
	DB         string         `json:"db"`
	Collection string         `json:"collection"`
	Filter     map[string]any `json:"filter"`
	Update     map[string]any `json:"update"`
	Multi      bool           `json:"multi"`
}

type DeleteRequest struct {
	DB         string         `json:"db"`
	Collection string         `json:"collection"`
	Filter     map[string]any `json:"filter"`
	Multi      bool           `json:"multi"`
}
type CreateIndexRequest struct {
	DB         string   `json:"db"`
	Collection string   `json:"collection"`

	// single field index (legacy support)
	Field string `json:"field"`

	// multi-field (compound) support
	Fields []string `json:"fields"`

	// "hash" or "btree"
	Type string `json:"type"`

	Unique     bool `json:"unique"`
	Background bool `json:"background"`
}
