package handlers

import "net/http"

func (h *Handlers) Databases(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"success": true, "databases": h.eng.Databases()})
}
