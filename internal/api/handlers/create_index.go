package handlers

import (
	"net/http"
	"strings"

	"testDB/internal/types"
)

func (h *Handlers) CreateIndex(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) { return }
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req types.CreateIndexRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}

	if req.DB == "" { req.DB = "default" }
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type == "" { req.Type = "hash" }

	// Support old API: field -> fields
	if len(req.Fields) == 0 && req.Field != "" {
		req.Fields = []string{req.Field}
	}
	if len(req.Fields) == 0 {
		writeJSON(w, 400, map[string]any{"success": false, "error": "field/fields required"})
		return
	}

	if err := h.eng.CreateIndex(req.DB, req.Collection, req.Fields, req.Type, req.Unique, req.Background); err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{"success": true})
}
