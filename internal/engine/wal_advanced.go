package engine

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type WALv2 struct {
	cfg         Config
	file        *os.File
	writer      *bufio.Writer
	mu          sync.Mutex
	sequence    int64
	batchCount  int
	lastSync    time.Time
	lastCheckpoint time.Time

	// Batch mode
	batchBuffer []*WALEntryV2
	batchMu     sync.Mutex
}

type WALEntryV2 struct {
	Sequence   int64          `json:"seq"`
	Timestamp  int64          `json:"ts"`
	Op         string         `json:"op"`
	DB         string         `json:"db"`
	Collection string         `json:"coll"`
	Doc        map[string]any `json:"doc,omitempty"`
	Filter     map[string]any `json:"filter,omitempty"`
	Update     map[string]any `json:"update,omitempty"`
	Multi      bool           `json:"multi,omitempty"`
	CRC        uint32         `json:"-"` // Computed, not stored in JSON
}

func NewWALv2(cfg Config) (*WALv2, error) {
	if err := os.MkdirAll(cfg.WALDir, 0755); err != nil {
		return nil, err
	}
	if cfg.EnableWALArchive {
		if err := os.MkdirAll(cfg.WALArchiveDir, 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(cfg.WALFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	wal := &WALv2{
		cfg:         cfg,
		file:        file,
		writer:      bufio.NewWriter(file),
		lastSync:    time.Now(),
		lastCheckpoint: time.Now(),
		batchBuffer: make([]*WALEntryV2, 0, cfg.WALBatchSize),
	}

	// Get current sequence
	wal.sequence = wal.getLastSequence()

	// Start background sync if async mode
	if cfg.WALSyncMode == WALSyncAsync {
		go wal.asyncSyncLoop()
	}

	// Start background batch flush if batch mode
	if cfg.WALSyncMode == WALSyncBatch {
		go wal.batchFlushLoop()
	}

	fmt.Printf("✅ WAL v2 initialized (mode: %s)\n", cfg.WALSyncMode)
	return wal, nil
}

// Append entry to WAL
func (w *WALv2) Append(entry WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.sequence++

	entryV2 := &WALEntryV2{
		Sequence:   w.sequence,
		Timestamp:  time.Now().UnixNano(),
		Op:         entry.Op,
		DB:         entry.DB,
		Collection: entry.Collection,
		Doc:        entry.Doc,
		Filter:     entry.Filter,
		Update:     entry.Update,
		Multi:      entry.Multi,
	}

	// Serialize to JSON
	data, err := json.Marshal(entryV2)
	if err != nil {
		return err
	}

	// Compute CRC
	entryV2.CRC = crc32.ChecksumIEEE(data)

	// Write in format: [length:4][data][crc:4]\n
	buf := new(bytes.Buffer)
	
	// Length
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	
	// Data
	buf.Write(data)
	
	// CRC
	if err := binary.Write(buf, binary.LittleEndian, entryV2.CRC); err != nil {
		return err
	}
	
	// Newline
	buf.WriteByte('\n')

	// Write to file
	if _, err := w.writer.Write(buf.Bytes()); err != nil {
		return err
	}

	w.batchCount++

	// Sync based on mode
	switch w.cfg.WALSyncMode {
	case WALSyncImmediate:
		return w.syncNow()
	case WALSyncBatch:
		if w.batchCount >= w.cfg.WALBatchSize {
			return w.syncNow()
		}
	case WALSyncAsync:
		// Will sync on timer
	}

	return nil
}

// Sync to disk
func (w *WALv2) syncNow() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	if err := w.file.Sync(); err != nil {
		return err
	}
	w.batchCount = 0
	w.lastSync = time.Now()
	return nil
}

// Async sync loop
func (w *WALv2) asyncSyncLoop() {
	ticker := time.NewTicker(w.cfg.WALBatchTimeout)
	defer ticker.Stop()

	for range ticker.C {
		w.mu.Lock()
		if w.batchCount > 0 {
			_ = w.syncNow()
		}
		w.mu.Unlock()
	}
}

// Batch flush loop
func (w *WALv2) batchFlushLoop() {
	ticker := time.NewTicker(w.cfg.WALBatchTimeout)
	defer ticker.Stop()

	for range ticker.C {
		w.mu.Lock()
		if w.batchCount > 0 && time.Since(w.lastSync) > w.cfg.WALBatchTimeout {
			_ = w.syncNow()
		}
		w.mu.Unlock()
	}
}

// Replay WAL entries
func (w *WALv2) Replay() ([]*WALEntryV2, error) {
	file, err := os.Open(w.cfg.WALFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	entries := make([]*WALEntryV2, 0)
	reader := bufio.NewReader(file)

	for {
		entry, err := w.readEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Corruption detected - stop here
			fmt.Printf("⚠️ WAL corruption at entry %d: %v\n", len(entries), err)
			break
		}
		entries = append(entries, entry)
	}

	fmt.Printf("✅ Replayed %d WAL entries\n", len(entries))
	return entries, nil
}

// Read single entry
func (w *WALv2) readEntry(reader *bufio.Reader) (*WALEntryV2, error) {
	// Read length
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}

	// Read CRC
	var storedCRC uint32
	if err := binary.Read(reader, binary.LittleEndian, &storedCRC); err != nil {
		return nil, err
	}

	// Skip newline
	reader.ReadByte()

	// Verify CRC
	computedCRC := crc32.ChecksumIEEE(data)
	if computedCRC != storedCRC {
		return nil, fmt.Errorf("CRC mismatch: expected %d, got %d", storedCRC, computedCRC)
	}

	// Decode JSON
	var entry WALEntryV2
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

// Checkpoint - save state and archive WAL
func (w *WALv2) Checkpoint() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Force sync first
	if err := w.syncNow(); err != nil {
		return err
	}

	// Close current WAL
	w.writer.Flush()
	w.file.Close()

	// Archive current WAL
	if w.cfg.EnableWALArchive {
		timestamp := time.Now().Format("20060102_150405")
		archivePath := filepath.Join(w.cfg.WALArchiveDir, fmt.Sprintf("wal_%s.log", timestamp))
		
		if err := os.Rename(w.cfg.WALFile, archivePath); err != nil {
			return err
		}
		
		fmt.Printf("📦 WAL archived: %s\n", archivePath)
	} else {
		// Just truncate
		if err := os.Remove(w.cfg.WALFile); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// Create new WAL file
	file, err := os.OpenFile(w.cfg.WALFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	w.file = file
	w.writer = bufio.NewWriter(file)
	w.lastCheckpoint = time.Now()

	fmt.Println("✅ Checkpoint completed")
	return nil
}

// Auto checkpoint loop
func (w *WALv2) StartAutoCheckpoint(onCheckpoint func() error) {
	go func() {
		ticker := time.NewTicker(w.cfg.CheckpointInterval)
		defer ticker.Stop()

		for range ticker.C {
			// Check if WAL is too large
			info, err := os.Stat(w.cfg.WALFile)
			if err != nil {
				continue
			}

			if info.Size() >= w.cfg.CheckpointWALSize {
				fmt.Println("🔄 Auto-checkpoint triggered (WAL size limit)")
				
				// Call external checkpoint callback (save all data)
				if onCheckpoint != nil {
					if err := onCheckpoint(); err != nil {
						fmt.Printf("❌ Checkpoint callback failed: %v\n", err)
						continue
					}
				}

				// Archive WAL
				if err := w.Checkpoint(); err != nil {
					fmt.Printf("❌ Checkpoint failed: %v\n", err)
				}
			}
		}
	}()

	fmt.Printf("✅ Auto-checkpoint started (every %v or %d bytes)\n", 
		w.cfg.CheckpointInterval, w.cfg.CheckpointWALSize)
}

// Get last sequence number
func (w *WALv2) getLastSequence() int64 {
	file, err := os.Open(w.cfg.WALFile)
	if err != nil {
		return 0
	}
	defer file.Close()

	var lastSeq int64
	reader := bufio.NewReader(file)

	for {
		entry, err := w.readEntry(reader)
		if err != nil {
			break
		}
		if entry.Sequence > lastSeq {
			lastSeq = entry.Sequence
		}
	}

	return lastSeq
}

// Close WAL
func (w *WALv2) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.syncNow(); err != nil {
		return err
	}

	return w.file.Close()
}

// Stats
func (w *WALv2) Stats() map[string]any {
	w.mu.Lock()
	defer w.mu.Unlock()

	info, _ := os.Stat(w.cfg.WALFile)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}

	return map[string]any{
		"sequence":       w.sequence,
		"size":           size,
		"batch_count":    w.batchCount,
		"last_sync":      w.lastSync,
		"last_checkpoint": w.lastCheckpoint,
		"sync_mode":      w.cfg.WALSyncMode,
	}
}