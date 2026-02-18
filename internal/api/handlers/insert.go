package handlers

import (
	"net/http"

	"testDB/internal/types"
)

func (h *Handlers) Insert(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) { return }
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req types.InsertRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}
	if req.DB == "" { req.DB = "default" }

	id, err := h.eng.Insert(req.DB, req.Collection, req.Data, true)
	if err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"success": true, "id": id})
}
