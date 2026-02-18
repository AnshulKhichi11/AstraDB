package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
	"testDB/internal/types"
)

type SegmentManager struct {
	dir           string
	segments      []*Segment
	activeSegment *Segment
	nextSegmentID int
	mu            sync.RWMutex
}


func NewSegmentManager(collectionDir string) (*SegmentManager, error) {
	sm := &SegmentManager{
		dir:      collectionDir,
		segments: make([]*Segment, 0),
	}

	// Load existing segments
	if err := sm.loadSegments(); err != nil {
		return nil, err
	}

	// Create active segment if needed
	if sm.activeSegment == nil {
		if err := sm.createNewSegment(); err != nil {
			return nil, err
		}
	}

	return sm, nil
}

func (sm *SegmentManager) loadSegments() error {
	segDir := filepath.Join(sm.dir, "segments")

	if err := os.MkdirAll(segDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(segDir)
	if err != nil {
		return err
	}

	segmentIDs := make([]int, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".seg" {
			var id int
			if _, err := fmt.Sscanf(entry.Name(), "%06d.seg", &id); err == nil {
				segmentIDs = append(segmentIDs, id)
			}
		}
	}

	sort.Ints(segmentIDs)

	for _, id := range segmentIDs {
		path := filepath.Join(segDir, fmt.Sprintf("%06d.seg", id))
		seg, err := OpenSegment(path, id)
		if err != nil {
			// Skip corrupted segments
			continue
		}

		sm.segments = append(sm.segments, seg)

		if id >= sm.nextSegmentID {
			sm.nextSegmentID = id + 1
		}

		// Last segment is active if not full
		if !seg.Sealed && seg.Size < SegmentSize {
			sm.activeSegment = seg
		} else {
			seg.Sealed = true
		}
	}

	return nil
}

func (sm *SegmentManager) createNewSegment() error {
	seg, err := NewSegment(sm.dir, sm.nextSegmentID)
	if err != nil {
		return err
	}

	sm.segments = append(sm.segments, seg)
	sm.activeSegment = seg
	sm.nextSegmentID++

	return nil
}

// Append document
func (sm *SegmentManager) Append(docID string, doc types.Document) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	rec := SegmentRecord{
		Type:  RecordInsert,
		DocID: docID,
		Data:  doc,
	}

	// Check if active segment is full
	if sm.activeSegment.Size >= SegmentSize {
		if err := sm.activeSegment.Seal(); err != nil {
			return err
		}
		if err := sm.createNewSegment(); err != nil {
			return err
		}
	}

	return sm.activeSegment.Append(rec)
}

// Read all documents
func (sm *SegmentManager) ReadAll() ([]types.Document, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Map of docID -> latest document
	docMap := make(map[string]types.Document)

	// Read all segments in order
	for _, seg := range sm.segments {
		records, err := seg.ReadAll()
		if err != nil {
			// Skip corrupted segments
			continue
		}

		for _, rec := range records {
			switch rec.Type {
			case RecordInsert, RecordUpdate:
				docMap[rec.DocID] = rec.Data
			case RecordDelete, RecordTombstone:
				delete(docMap, rec.DocID)
			}
		}
	}

	// Convert to slice
	docs := make([]types.Document, 0, len(docMap))
	for _, doc := range docMap {
		docs = append(docs, doc)
	}

	return docs, nil
}

// Delete document (append tombstone)
func (sm *SegmentManager) Delete(docID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	rec := SegmentRecord{
		Type:  RecordTombstone,
		DocID: docID,
		Data:  nil,
	}

	// Check if active segment is full
	if sm.activeSegment.Size >= SegmentSize {
		if err := sm.activeSegment.Seal(); err != nil {
			return err
		}
		if err := sm.createNewSegment(); err != nil {
			return err
		}
	}

	return sm.activeSegment.Append(rec)
}

// Close all segments
func (sm *SegmentManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, seg := range sm.segments {
		if err := seg.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Compact segments (merge and remove tombstones)
func (sm *SegmentManager) Compact() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Don't compact if only 1 segment
	if len(sm.segments) <= 1 {
		return nil
	}

	// Build latest state (skip tombstones)
	docMap := make(map[string]types.Document)

	for _, seg := range sm.segments[:len(sm.segments)-1] { // exclude active
		records, err := seg.ReadAll()
		if err != nil {
			continue
		}

		for _, rec := range records {
			switch rec.Type {
			case RecordInsert, RecordUpdate:
				docMap[rec.DocID] = rec.Data
			case RecordDelete, RecordTombstone:
				delete(docMap, rec.DocID)
			}
		}
	}

	// Create new compacted segment
	compactedID := sm.nextSegmentID
	compacted, err := NewSegment(sm.dir, compactedID)
	if err != nil {
		return err
	}
	sm.nextSegmentID++

	// Write all live documents
	for docID, doc := range docMap {
		rec := SegmentRecord{
			Type:  RecordInsert,
			DocID: docID,
			Data:  doc,
		}
		if err := compacted.Append(rec); err != nil {
			compacted.Close()
			os.Remove(compacted.Path)
			return err
		}
	}

	compacted.Seal()

	// Delete old segments
	for _, seg := range sm.segments[:len(sm.segments)-1] {
		seg.Close()
		os.Remove(seg.Path)
	}

	// Update segment list
	sm.segments = []*Segment{compacted, sm.activeSegment}

	return nil
}




// StartAutoCompaction runs compaction in background
func (sm *SegmentManager) StartAutoCompaction(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			sm.mu.RLock()
			segCount := len(sm.segments)
			sm.mu.RUnlock()

			if segCount >= 3 {
				fmt.Printf("🔄 Auto-compaction starting (%d segments)...\n", segCount)
				
				start := time.Now()
				if err := sm.Compact(); err != nil {
					fmt.Printf("❌ Auto-compaction failed: %v\n", err)
				} else {
					duration := time.Since(start)
					fmt.Printf("✅ Auto-compaction completed in %v\n", duration)
				}
			}
		}
	}()
}

// GetStats returns segment statistics
func (sm *SegmentManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	totalSize := int64(0)
	totalDocs := 0
	segmentInfo := make([]map[string]interface{}, 0)

	for _, seg := range sm.segments {
		seg.mu.RLock()
		info := map[string]interface{}{
			"id":     seg.ID,
			"size":   seg.Size,
			"docs":   seg.DocCount,
			"sealed": seg.Sealed,
		}
		seg.mu.RUnlock()

		segmentInfo = append(segmentInfo, info)
		totalSize += seg.Size
		totalDocs += seg.DocCount
	}

	return map[string]interface{}{
		"segment_count": len(sm.segments),
		"total_size":    totalSize,
		"total_docs":    totalDocs,
		"segments":      segmentInfo,
	}
}