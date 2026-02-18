package handlers

import "testDB/internal/engine"

type Handlers struct {
	eng *engine.Engine
}

func New(e *engine.Engine) *Handlers {
	return &Handlers{eng: e}
}
