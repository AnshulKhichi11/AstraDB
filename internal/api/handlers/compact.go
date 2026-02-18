package handlers

import (
	"net/http"
)

func (h *Handlers) Compact(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	db := r.URL.Query().Get("db")
	coll := r.URL.Query().Get("collection")

	if db == "" {
		db = "default"
	}

	if coll == "" {
		writeJSON(w, 400, map[string]any{
			"success": false,
			"error":   "collection parameter required",
		})
		return
	}

	if err := h.eng.CompactCollection(db, coll); err != nil {
		writeJSON(w, 500, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, 200, map[string]any{
		"success": true,
		"message": "Compaction completed",
	})
}

func (h *Handlers) SegmentStats(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	db := r.URL.Query().Get("db")
	stats := h.eng.GetSegmentStats(db)

	writeJSON(w, 200, map[string]any{
		"success": true,
		"stats":   stats,
	})
}