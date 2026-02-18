package handlers

import (
	"net/http"
	"time"
)

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "time": time.Now().Unix()})
}
