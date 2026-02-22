package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"testDB/internal/types"
)

func (c *Collection) saveLocked() error {
	// Deduplicate by _id (preserve last occurrence)
	seen := map[string]bool{}
	uniqueRev := make([]types.Document, 0, len(c.Docs))
	for i := len(c.Docs) - 1; i >= 0; i-- {
		d := c.Docs[i]
		idStr := ""
		if id, ok := d["_id"].(string); ok {
			idStr = id
		}
		if idStr == "" {
			uniqueRev = append(uniqueRev, d)
			continue
		}
		if _, ok := seen[idStr]; ok {
			continue
		}
		seen[idStr] = true
		uniqueRev = append(uniqueRev, d)
	}
	unique := make([]types.Document, 0, len(uniqueRev))
	for i := len(uniqueRev) - 1; i >= 0; i-- {
		unique = append(unique, uniqueRev[i])
	}

	b, err := json.MarshalIndent(unique, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(c.DataFile, b, 0666)
}

func (e *Engine) flushAll() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, db := range e.databases {
		db.mu.RLock()
		for _, c := range db.collections {
			c.mu.Lock()
			_ = c.saveLocked()
			c.mu.Unlock()
		}
		db.mu.RUnlock()
	}
	return nil
}

func (e *Engine) loadSnapshots() error {
	entries, err := os.ReadDir(e.cfg.DBsDir)
	if err != nil {
		return err
	}

	for _, dbEnt := range entries {
		if !dbEnt.IsDir() {
			continue
		}

		dbName := dbEnt.Name()
		db, _ := e.getOrCreateDB(dbName)

		collsPath := filepath.Join(e.cfg.DBsDir, dbName, "collections")
		collDirs, err := os.ReadDir(collsPath)
		if err != nil {
			continue
		}

		for _, cEnt := range collDirs {
			if !cEnt.IsDir() {
				continue
			}

			cName := cEnt.Name()
			collDir := filepath.Join(collsPath, cName)
			dataFile := filepath.Join(collsPath, cName, "data.db")

			// ── CRITICAL FIX ──────────────────────────────────────────────
			// Must initialize SegmentManager here, exactly like
			// getOrCreateCollection() does. Previously this was missing,
			// so useSegments was always false during startup, which meant
			// the duplicate-guard in Insert() never triggered during WAL
			// replay → every doc got inserted twice on restart.
			segMgr, segErr := NewSegmentManager(collDir)
			useSegs := (segErr == nil)

			var docs []types.Document

			if useSegs {
				// Segments are the source of truth.
				// Do NOT load data.db too — that would create duplicates.
				// Docs are read lazily via segmentMgr.ReadAll().
				docs = []types.Document{}
			} else {
				// Legacy path: no segments, load from JSON data.db
				if b, err := os.ReadFile(dataFile); err == nil && len(b) > 0 {
					dec := json.NewDecoder(bytes.NewReader(b))
					dec.UseNumber()

					if decodeErr := dec.Decode(&docs); decodeErr != nil {
						docs = []types.Document{}
					} else {
						for i := range docs {
							docs[i] = canonicalizeDocument(docs[i])
						}
						// Deduplicate by _id (preserve last occurrence)
						seen := map[string]bool{}
						uniqueRev := make([]types.Document, 0, len(docs))
						for i := len(docs) - 1; i >= 0; i-- {
							d := docs[i]
							idStr := ""
							if id, ok := d["_id"].(string); ok {
								idStr = id
							}
							if idStr == "" {
								uniqueRev = append(uniqueRev, d)
								continue
							}
							if _, ok := seen[idStr]; ok {
								continue
							}
							seen[idStr] = true
							uniqueRev = append(uniqueRev, d)
						}
						uniq := make([]types.Document, 0, len(uniqueRev))
						for i := len(uniqueRev) - 1; i >= 0; i-- {
							uniq = append(uniq, uniqueRev[i])
						}
						docs = uniq
					}
				}
			}

			col := &Collection{
				Name:        cName,
				DataFile:    dataFile,
				Docs:        docs,
				segmentMgr:  segMgr,  // ← NOW CORRECTLY SET
				useSegments: useSegs, // ← NOW CORRECTLY SET
			}

			col.loadIndexMetas(e.cfg)
			db.collections[cName] = col
		}
	}

	return nil
}
