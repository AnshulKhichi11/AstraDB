package handlers

import (
	"net/http"

	"testDB/internal/types"
)

func (h *Handlers) Query(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) { return }
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]any{"success": false, "error": "Method not allowed"})
		return
	}

	var req types.QueryRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": "Invalid JSON: " + err.Error()})
		return
	}
	if req.DB == "" { req.DB = "default" }

	res, err := h.eng.Query(req.DB, req.Collection, req.Filter, req.Sort, req.Limit, req.Skip, req.Projection)
	if err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"success": true, "count": len(res), "data": res})
}
