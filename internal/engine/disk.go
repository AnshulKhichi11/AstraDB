package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"testDB/internal/types"
)

func (c *Collection) saveLocked() error {
	b, err := json.MarshalIndent(c.Docs, "", "  ")
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
			dataFile := filepath.Join(collsPath, cName, "data.db")

			var docs []types.Document

			// Load documents if the file exists and has content
			if b, err := os.ReadFile(dataFile); err == nil && len(b) > 0 {
				dec := json.NewDecoder(bytes.NewReader(b))
				dec.UseNumber() // preserve numeric types (int vs float64)

				if decodeErr := dec.Decode(&docs); decodeErr != nil {
					// You can add logging here in real code
					// log.Printf("failed to decode %s: %v", dataFile, decodeErr)
					docs = []types.Document{} // fallback to empty slice
				} else {
					// Normalize / canonicalize each document
					for i := range docs {
						docs[i] = canonicalizeDocument(docs[i])
					}
				}
			} else if err != nil && !os.IsNotExist(err) {
				// Optional: log unexpected read errors
				// log.Printf("failed to read %s: %v", dataFile, err)
			}

			// ───────────────────────────────────────────────────────────────
			// Create collection → load indexes → register collection
			// This is the recommended and cleanest place
			// ───────────────────────────────────────────────────────────────
			col := &Collection{
				Name:     cName,
				DataFile: dataFile,
				Docs:     docs,
			}

			// Load index metadata right after creating the collection
			col.loadIndexMetas(e.cfg)

			// Now it's fully initialized → safe to put in the map
			db.collections[cName] = col
		}
	}

	return nil
}