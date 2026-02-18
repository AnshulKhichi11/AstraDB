package handlers

import "net/http"

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	writeJSON(w, 200, map[string]any{"success": true, "stats": h.eng.Stats(db)})
}
