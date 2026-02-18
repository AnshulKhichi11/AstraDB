package engine

import (
	"fmt"
    "os"
    "testing"

    "testDB/internal/types"
)

func TestSegmentBasics(t *testing.T) {
    dir := "./test_segments"
    os.RemoveAll(dir)
    defer os.RemoveAll(dir)

    // Create segment
    seg, err := NewSegment(dir, 1)
    if err != nil {
        t.Fatal(err)
    }
    defer seg.Close()

    // Append records
    for i := 0; i < 100; i++ {
        doc := types.Document{
            "_id":  fmt.Sprintf("doc%d", i),
            "name": fmt.Sprintf("User%d", i),
            "age":  20 + i,
        }

        rec := SegmentRecord{
            Type:  RecordInsert,
            DocID: fmt.Sprintf("doc%d", i),
            Data:  doc,
        }

        if err := seg.Append(rec); err != nil {
            t.Fatal(err)
        }
    }

    // Read back
    records, err := seg.ReadAll()
    if err != nil {
        t.Fatal(err)
    }

    if len(records) != 100 {
        t.Fatalf("Expected 100 records, got %d", len(records))
    }

    // Verify data
    if records[0].Data["name"] != "User0" {
        t.Fatalf("Wrong data: %v", records[0].Data)
    }
}

func TestCompaction(t *testing.T) {
    dir := "./test_compact"
    os.RemoveAll(dir)
    defer os.RemoveAll(dir)

    sm, err := NewSegmentManager(dir)
    if err != nil {
        t.Fatal(err)
    }
    defer sm.Close()

    // Insert 100 docs
    for i := 0; i < 100; i++ {
        doc := types.Document{
            "_id":  fmt.Sprintf("doc%d", i),
            "name": fmt.Sprintf("User%d", i),
        }
        if err := sm.Append(fmt.Sprintf("doc%d", i), doc); err != nil {
            t.Fatal(err)
        }
    }

    // Delete 50 docs
    for i := 0; i < 50; i++ {
        if err := sm.Delete(fmt.Sprintf("doc%d", i)); err != nil {
            t.Fatal(err)
        }
    }

    // Compact
    if err := sm.Compact(); err != nil {
        t.Fatal(err)
    }

    // Verify only 50 docs remain
    docs, err := sm.ReadAll()
    if err != nil {
        t.Fatal(err)
    }

    if len(docs) != 50 {
        t.Fatalf("Expected 50 docs after compaction, got %d", len(docs))
    }
}