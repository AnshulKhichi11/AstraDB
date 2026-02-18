package handlers

import "net/http"

func (h *Handlers) Collections(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	if db == "" { db = "default" }

	colls, err := h.eng.Collections(db)
	if err != nil {
		writeJSON(w, 400, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"success": true, "db": db, "collections": colls})
}
