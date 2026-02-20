package engine

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"testDB/internal/types"
)

type Engine struct {
	mu        sync.RWMutex
	databases map[string]*Database

	walMu     sync.Mutex
	replaying bool

	cfg Config
	walv2 *WALv2 
}

type Database struct {
	Name        string
	mu          sync.RWMutex
	collections map[string]*Collection
}

type Collection struct {
	Name     string
	DataFile string
	mu       sync.RWMutex
	Docs     []types.Document // Keep for backward compatibility
	
	// NEW: Segment-based storage
	segmentMgr  *SegmentManager
	useSegments bool
	
	Indexes      map[string]*Index
	IndexesHash  map[string]*HashIndex
	IndexesBTree map[string]*BTreeIndex
	IndexMetas   map[string]IndexMeta
}

type WALEntry struct {
	TS         int64          `json:"ts"`
	Op         string         `json:"op"` // insert|update|delete
	DB         string         `json:"db"`
	Collection string         `json:"collection"`
	Doc        types.Document `json:"doc,omitempty"`
	Filter     map[string]any `json:"filter,omitempty"`
	Update     map[string]any `json:"update,omitempty"`
	Multi      bool           `json:"multi,omitempty"`
}


func New(cfg Config) (*Engine, error) {
	if err := ensureDirs(cfg); err != nil {
		return nil, err
	}

	e := &Engine{
		databases: map[string]*Database{},
		cfg:       cfg,
	}

	// Initialize WAL v2
	walv2, err := NewWALv2(cfg)
	if err != nil {
		return nil, err
	}
	e.walv2 = walv2

	// Load snapshots first
	if err := e.loadSnapshots(); err != nil {
		return nil, err
	}

	// Replay WAL
	if err := e.replayWALv2(); err != nil {
		return nil, err
	}

	// Start auto-checkpoint
	walv2.StartAutoCheckpoint(func() error {
    if err := e.flushAll(); err != nil {
        return err
    }
    return e.walv2.Checkpoint() // ✅ clear WAL after durable snapshot
})

	return e, nil
}

func (e *Engine) getOrCreateDB(dbName string) (*Database, error) {
	dbName = normalizeName(dbName)
	if dbName == "" {
		return nil, errors.New("db is required")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if db, ok := e.databases[dbName]; ok {
		return db, nil
	}

	if err := mkdirAll(dbCollectionsDir(e.cfg, dbName)); err != nil {
		return nil, err
	}

	db := &Database{Name: dbName, collections: map[string]*Collection{}}
	e.databases[dbName] = db
	return db, nil
}

func (db *Database) getOrCreateCollection(cfg Config, collName string) (*Collection, error) {
	collName = normalizeName(collName)
	if collName == "" {
		return nil, errors.New("collection is required")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if c, ok := db.collections[collName]; ok {
		return c, nil
	}

	collDir := collectionDir(cfg, db.Name, collName)
	if err := mkdirAll(collDir); err != nil {
		return nil, err
	}

	dataFile := collectionDataFile(cfg, db.Name, collName)

	// Try to create segment manager
	segMgr, segErr := NewSegmentManager(collDir)
	useSegs := (segErr == nil)

	c := &Collection{
		Name:        collName,
		DataFile:    dataFile,
		Docs:        []types.Document{},
		segmentMgr:  segMgr,
		useSegments: useSegs,
	}
	c.loadIndexMetas(cfg)
	db.collections[collName] = c

	if !useSegs {
		_ = c.saveLocked() // create empty file (old way)
	}
	return c, nil
}

func (e *Engine) Insert(dbName, collName string, doc types.Document, doLog bool) (string, error) {
	db, err := e.getOrCreateDB(dbName)
	if err != nil {
		return "", err
	}
	c, err := db.getOrCreateCollection(e.cfg, collName)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", errors.New("data is required")
	}

	doc = canonicalizeDocument(doc)

	// A1: Document size limit
	if err := enforceDocSizeLimit(doc, e.cfg.MaxDocBytes); err != nil {
		return "", err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := doc["_id"]; !ok {
		doc["_id"] = NewObjectID() // A1: ObjectId-like sortable id
	}
	if _, ok := doc["_created"]; !ok {
		doc["_created"] = time.Now().Unix()
	}

	docID := fmt.Sprintf("%v", doc["_id"])

	// Use segments if available, otherwise fallback to old method
	if c.useSegments && c.segmentMgr != nil {
		if err := c.segmentMgr.Append(docID, doc); err != nil {
			return "", err
		}
	} else {
		// Old method (backward compatibility)
		c.Docs = append(c.Docs, doc)
		if err := c.saveLocked(); err != nil {
			return "", err
		}
	}

	for _, idx := range c.Indexes {
		val := getIndexValue(doc, idx.Field)
		if idx.Unique && len(idx.Entries[val]) > 0 {
			return "", fmt.Errorf("duplicate value for unique index: %s", idx.Field)
		}
		idx.Entries[val] = append(idx.Entries[val], docID)
	}

	if doLog && !e.replaying {
		_ = e.walAppend(WALEntry{TS: time.Now().Unix(), Op: "insert", DB: db.Name, Collection: c.Name, Doc: doc})
	}

	return docID, nil
}

func (e *Engine) Query(
	dbName, collName string,
	filter map[string]any,
	sortSpec map[string]int,
	limit, skip int,
	projection map[string]int,
) ([]types.Document, error) {

	db, err := e.getOrCreateDB(dbName)
	if err != nil {
		return nil, err
	}

	c, err := db.getOrCreateCollection(e.cfg, collName)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var out []types.Document

	// ────────────────────────────────────────────────────────────────
	// Fast path: try to use index to reduce scanned documents
	// ────────────────────────────────────────────────────────────────
	if ids, ok := c.candidateIDsByIndex(filter); ok && len(ids) > 0 {

		// Get all docs (from segments or memory)
		var allDocs []types.Document
		if c.useSegments && c.segmentMgr != nil {
			var err error
			allDocs, err = c.segmentMgr.ReadAll()
			if err != nil {
				return nil, err
			}
		} else {
			allDocs = c.Docs
		}

		// Create ID set for fast lookup
		idSet := make(map[string]bool)
		for _, id := range ids {
			idSet[id] = true
		}

		out = make([]types.Document, 0, len(ids))
		for _, d := range allDocs {
			if docID, ok := d["_id"].(string); ok && idSet[docID] {
				// Safety / correctness: re-check full filter
				if matchesFilter(d, filter) {
					out = append(out, d)
				}
			}
		}

	} else {
		// Slow path: full collection scan
		var allDocs []types.Document

		if c.useSegments && c.segmentMgr != nil {
			// Read from segments
			var err error
			allDocs, err = c.segmentMgr.ReadAll()
			if err != nil {
				return nil, err
			}
		} else {
			// Old method: use in-memory Docs
			allDocs = c.Docs
		}

		out = make([]types.Document, 0, len(allDocs)/4)
		for _, d := range allDocs {
			if matchesFilter(d, filter) {
				out = append(out, d)
			}
		}
	}

	// ────────────────────────────────────────────────────────────────
	// Common post-filter steps: sort → skip/limit → projection
	// ────────────────────────────────────────────────────────────────
	applySort(out, sortSpec)

	out = applySkipLimit(out, skip, limit)

	// Apply projection (field inclusion/exclusion)
	out = applyProjection(out, projection)

	// Return a shallow copy (defensive — prevents caller from mutating internal data)
	res := make([]types.Document, len(out))
	copy(res, out)

	return res, nil
}

func (e *Engine) Update(dbName, collName string, filter map[string]any, update map[string]any, multi bool, doLog bool) (int, error) {
	db, err := e.getOrCreateDB(dbName)
	if err != nil {
		return 0, err
	}
	c, err := db.getOrCreateCollection(e.cfg, collName)
	if err != nil {
		return 0, err
	}
	if update == nil {
		return 0, errors.New("update is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Get all docs
	var allDocs []types.Document
	if c.useSegments && c.segmentMgr != nil {
		var err error
		allDocs, err = c.segmentMgr.ReadAll()
		if err != nil {
			return 0, err
		}
	} else {
		allDocs = c.Docs
	}

	updated := 0
	for i := range allDocs {
		d := allDocs[i]
		if matchesFilter(d, filter) {
			if err := applyUpdateOperators(d, update); err != nil {
				return 0, err
			}
			d["_updated"] = time.Now().Unix()

			// A1 doc size limit after update
			if err := enforceDocSizeLimit(d, e.cfg.MaxDocBytes); err != nil {
				return 0, err
			}

			// If using segments, append updated doc
			if c.useSegments && c.segmentMgr != nil {
				docID := fmt.Sprintf("%v", d["_id"])
				if err := c.segmentMgr.Append(docID, d); err != nil {
					return 0, err
				}
			}

			updated++
			if !multi {
				break
			}
		}
	}

	if updated > 0 {
		if c.useSegments && c.segmentMgr != nil {
			// Segment storage handles updates via append
			// (already appended in the loop)
		} else {
			// Old method
			c.Docs = allDocs // Update in-memory
			if err := c.saveLocked(); err != nil {
				return 0, err
			}
		}
		if doLog && !e.replaying {
			_ = e.walAppend(WALEntry{TS: time.Now().Unix(), Op: "update", DB: db.Name, Collection: c.Name, Filter: filter, Update: update, Multi: multi})
		}
	}
	return updated, nil
}

func (e *Engine) Delete(dbName, collName string, filter map[string]any, multi bool, doLog bool) (int, error) {
	db, err := e.getOrCreateDB(dbName)
	if err != nil {
		return 0, err
	}
	c, err := db.getOrCreateCollection(e.cfg, collName)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Get all docs
	var allDocs []types.Document
	if c.useSegments && c.segmentMgr != nil {
		var err error
		allDocs, err = c.segmentMgr.ReadAll()
		if err != nil {
			return 0, err
		}
	} else {
		allDocs = c.Docs
	}

	deleted := 0
	var newDocs []types.Document

	if !c.useSegments {
		newDocs = make([]types.Document, 0, len(allDocs))
	}

	for _, d := range allDocs {
		if matchesFilter(d, filter) {
			// Delete from segments (tombstone)
			if c.useSegments && c.segmentMgr != nil {
				docID := fmt.Sprintf("%v", d["_id"])
				if err := c.segmentMgr.Delete(docID); err != nil {
					return 0, err
				}
			}

			deleted++
			if !multi && deleted == 1 {
				if !c.useSegments {
					continue
				} else {
					break
				}
			}
			if multi {
				if !c.useSegments {
					continue
				}
			}
		}

		if !c.useSegments {
			newDocs = append(newDocs, d)
		}
	}

	if deleted > 0 {
		if !c.useSegments {
			c.Docs = newDocs
			if err := c.saveLocked(); err != nil {
				return 0, err
			}
		}
		if doLog && !e.replaying {
			_ = e.walAppend(WALEntry{TS: time.Now().Unix(), Op: "delete", DB: db.Name, Collection: c.Name, Filter: filter, Multi: multi})
		}
	}

	return deleted, nil
}

func (e *Engine) Databases() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return mapKeysSorted(e.databases)
}

func (e *Engine) Collections(dbName string) ([]string, error) {
	db, err := e.getOrCreateDB(dbName)
	if err != nil {
		return nil, err
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	return mapKeysSorted(db.collections), nil
}

func (e *Engine) Stats(dbName string) map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()

	totalDB := len(e.databases)
	totalColl := 0
	totalDocs := 0

	if dbName != "" {
		db, ok := e.databases[normalizeName(dbName)]
		if !ok {
			return map[string]any{"databases": 0, "collections": 0, "documents": 0}
		}
		db.mu.RLock()
		defer db.mu.RUnlock()
		for _, c := range db.collections {
			totalColl++
			c.mu.RLock()
			if c.useSegments && c.segmentMgr != nil {
				docs, _ := c.segmentMgr.ReadAll()
				totalDocs += len(docs)
			} else {
				totalDocs += len(c.Docs)
			}
			c.mu.RUnlock()
		}
		return map[string]any{"db": dbName, "collections": totalColl, "documents": totalDocs}
	}

	for _, db := range e.databases {
		db.mu.RLock()
		totalColl += len(db.collections)
		for _, c := range db.collections {
			c.mu.RLock()
			if c.useSegments && c.segmentMgr != nil {
				docs, _ := c.segmentMgr.ReadAll()
				totalDocs += len(docs)
			} else {
				totalDocs += len(c.Docs)
			}
			c.mu.RUnlock()
		}
		db.mu.RUnlock()
	}

	return map[string]any{"databases": totalDB, "collections": totalColl, "documents": totalDocs}
}

// Shutdown closes all segments cleanly
func (e *Engine) Shutdown() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	fmt.Println("Shutting down AstraDB...")

	for _, db := range e.databases {
		db.mu.RLock()
		for _, c := range db.collections {
			c.mu.Lock()
			if c.useSegments && c.segmentMgr != nil {
				fmt.Printf("Compacting %s/%s\n", db.Name, c.Name)
				_ = c.segmentMgr.Compact()
				_ = c.segmentMgr.Close()
			}
			c.mu.Unlock()
		}
		db.mu.RUnlock()
	}
	_ = e.flushAll()
_ = e.walv2.Checkpoint()

	fmt.Println("Shutdown complete")
	return nil
}



// StartAutoCompaction starts background compaction for all collections
func (e *Engine) StartAutoCompaction(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			e.mu.RLock()
			databases := make([]*Database, 0, len(e.databases))
			for _, db := range e.databases {
				databases = append(databases, db)
			}
			e.mu.RUnlock()

			for _, db := range databases {
				db.mu.RLock()
				collections := make([]*Collection, 0, len(db.collections))
				for _, c := range db.collections {
					collections = append(collections, c)
				}
				db.mu.RUnlock()

				for _, c := range collections {
					c.mu.Lock()
					if c.useSegments && c.segmentMgr != nil {
						if len(c.segmentMgr.segments) >= 3 {
							_ = c.segmentMgr.Compact()
						}
					}
					c.mu.Unlock()
				}
			}
		}
	}()

	fmt.Printf("✅ Auto-compaction started (every %v)\n", interval)
}

// CompactCollection manually triggers compaction
func (e *Engine) CompactCollection(dbName, collName string) error {
	db, err := e.getOrCreateDB(dbName)
	if err != nil {
		return err
	}

	c, err := db.getOrCreateCollection(e.cfg, collName)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.useSegments && c.segmentMgr != nil {
		return c.segmentMgr.Compact()
	}

	return nil
}

// GetSegmentStats returns segment statistics
func (e *Engine) GetSegmentStats(dbName string) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := make(map[string]interface{})

	if dbName != "" {
		db, ok := e.databases[normalizeName(dbName)]
		if !ok {
			return stats
		}

		db.mu.RLock()
		defer db.mu.RUnlock()

		collStats := make(map[string]interface{})
		for name, c := range db.collections {
			c.mu.RLock()
			if c.useSegments && c.segmentMgr != nil {
				collStats[name] = c.segmentMgr.GetStats()
			}
			c.mu.RUnlock()
		}
		stats["collections"] = collStats
	}

	return stats
}