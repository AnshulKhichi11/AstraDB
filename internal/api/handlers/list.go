package handlers

import "net/http"

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	if db == "" { db = "default" }
	coll := r.URL.Query().Get("collection")

	res, err := h.eng.Query(db, coll, map[string]any{}, map[string]int{}, 0, 0, nil)
	if err != nil {
		writeJSON(w, 500, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"success": true, "count": len(res), "data": res})
}
