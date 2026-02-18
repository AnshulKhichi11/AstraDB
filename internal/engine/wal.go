package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
)


func (e *Engine) walAppend(entry WALEntry) error {
	if e.replaying {
		return nil
	}
	return e.walv2.Append(entry)
}

func (e *Engine) replayWAL() error {
	e.walMu.Lock()
	defer e.walMu.Unlock()

	f, err := os.Open(e.cfg.WALFile)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 8*1024*1024)

	e.replaying = true
	defer func() { e.replaying = false }()

	applied := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		dec := json.NewDecoder(bytes.NewReader([]byte(line)))
		dec.UseNumber()

		var ent WALEntry
		if err := dec.Decode(&ent); err != nil {
			continue
		}

		switch ent.Op {
		case "insert":
			_, _ = e.Insert(ent.DB, ent.Collection, ent.Doc, false)
		case "update":
			_, _ = e.Update(ent.DB, ent.Collection, ent.Filter, ent.Update, ent.Multi, false)
		case "delete":
			_, _ = e.Delete(ent.DB, ent.Collection, ent.Filter, ent.Multi, false)
		}
		applied++
	}

	if err := sc.Err(); err != nil {
		return err
	}

	// simple checkpoint (we will harden in A5 later)
	_ = e.flushAll()
	_ = os.WriteFile(e.cfg.WALFile, []byte(""), 0666)

	return nil
}



func (e *Engine) replayWALv2() error {
	e.walMu.Lock()
	e.replaying = true
	defer func() {
		e.replaying = false
		e.walMu.Unlock()
	}()

	entries, err := e.walv2.Replay()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Op {
		case "insert":
			_, _ = e.Insert(entry.DB, entry.Collection, entry.Doc, false)
		case "update":
			_, _ = e.Update(entry.DB, entry.Collection, entry.Filter, entry.Update, entry.Multi, false)
		case "delete":
			_, _ = e.Delete(entry.DB, entry.Collection, entry.Filter, entry.Multi, false)
		}
	}

	// After replay, checkpoint (clear WAL)
	if len(entries) > 0 {
		if err := e.flushAll(); err != nil {
			return err
		}
		if err := e.walv2.Checkpoint(); err != nil {
			return err
		}
	}

	return nil
}


func (e *Engine) WALStats() map[string]any {

	e.walMu.Lock()
	defer e.walMu.Unlock()

	info, err := os.Stat(e.cfg.WALFile)
	if err != nil {
		return map[string]any{
			"size":  0,
			"error": err.Error(),
		}
	}

	// count entries
	file, err := os.Open(e.cfg.WALFile)
	if err != nil {
		return map[string]any{
			"size": info.Size(),
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		count++
	}

	return map[string]any{
		"file":    e.cfg.WALFile,
		"size":    info.Size(),
		"entries": count,
	}
}
