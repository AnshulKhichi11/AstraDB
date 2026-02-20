package handlers

import (
	"net/http"
	"time"

	"testDB/internal/engine"
)

// SaveSchema handles creating a new schema for a collection
func (h *Handlers) SaveSchema(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		DB         string                        `json:"db"`
		Collection string                        `json:"collection"`
		Fields     map[string]engine.FieldSchema `json:"fields"`
		Strict     bool                          `json:"strict"`
	}
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}

	schema := &engine.Schema{
		Version:   1,
		Fields:    req.Fields,
		Strict:    req.Strict,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.eng.SaveSchema(req.DB, req.Collection, schema); err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{"success": true})
}

// GetSchema returns the stored schema for a collection
func (h *Handlers) GetSchema(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}
	if r.Method != "GET" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	db := r.URL.Query().Get("db")
	collection := r.URL.Query().Get("collection")
	if db == "" || collection == "" {
		writeJSON(w, 400, map[string]any{"success": false, "error": "db and collection required"})
		return
	}

	schema, err := h.eng.LoadSchema(db, collection)
	if err != nil {
		writeJSON(w, 404, map[string]any{"success": false, "error": "Schema not found"})
		return
	}
	writeJSON(w, 200, schema)
}

// UpdateSchema updates (or creates) a schema for a collection
func (h *Handlers) UpdateSchema(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		DB         string                        `json:"db"`
		Collection string                        `json:"collection"`
		Fields     map[string]engine.FieldSchema `json:"fields"`
		Strict     bool                          `json:"strict"`
	}
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}

	// Preserve createdAt when possible
	existing, _ := h.eng.LoadSchema(req.DB, req.Collection)
	createdAt := time.Now()
	if existing != nil {
		createdAt = existing.CreatedAt
	}

	schema := &engine.Schema{
		Version:   1,
		Fields:    req.Fields,
		Strict:    req.Strict,
		CreatedAt: createdAt,
		UpdatedAt: time.Now(),
	}

	if err := h.eng.SaveSchema(req.DB, req.Collection, schema); err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"success": true})
}

// DeleteSchema removes the schema file for a collection
func (h *Handlers) DeleteSchema(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		DB         string `json:"db"`
		Collection string `json:"collection"`
	}
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}

	if err := h.eng.DeleteSchema(req.DB, req.Collection); err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"success": true})
}

// ValidateDocument validates a document against a schema without inserting
func (h *Handlers) ValidateDocument(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req struct {
		DB         string                 `json:"db"`
		Collection string                 `json:"collection"`
		Document   map[string]interface{} `json:"document"`
	}
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}

	schema, err := h.eng.LoadSchema(req.DB, req.Collection)
	if err != nil || schema == nil {
		writeJSON(w, 200, map[string]any{"valid": true, "message": "No schema defined, validation skipped"})
		return
	}

	result := schema.Validate(req.Document)
	writeJSON(w, 200, result)
}
