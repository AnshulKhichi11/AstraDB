package handlers

import (
	"net/http"
)

func (h *Handlers) WALStats(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	stats := h.eng.WALStats()

	writeJSON(w, 200, map[string]any{
		"success": true,
		"stats":   stats,
	})
}