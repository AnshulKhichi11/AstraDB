package engine

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"

	"testDB/internal/types"
)

const (
	SegmentSize    = 10 * 1024 * 1024 // 10MB per segment
	SegmentMagic   = uint32(0x41535447) // "ASTG"
	SegmentVersion = uint32(1)
)

type RecordType byte

const (
	RecordInsert    RecordType = 0
	RecordUpdate    RecordType = 1
	RecordDelete    RecordType = 2
	RecordTombstone RecordType = 3
)

type SegmentRecord struct {
	Type   RecordType
	DocID  string
	Data   types.Document
	Offset int64
}

type Segment struct {
	ID       int
	Path     string
	Size     int64
	DocCount int
	Sealed   bool
	mu       sync.RWMutex
	file     *os.File
}

// Create new segment
func NewSegment(dir string, id int) (*Segment, error) {
	segDir := filepath.Join(dir, "segments")
	if err := os.MkdirAll(segDir, 0755); err != nil {
		return nil, err
	}

	path := filepath.Join(segDir, fmt.Sprintf("%06d.seg", id))

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	seg := &Segment{
		ID:       id,
		Path:     path,
		file:     file,
		DocCount: 0,
		Sealed:   false,
	}

	// Write header
	if err := seg.writeHeader(); err != nil {
		file.Close()
		return nil, err
	}

	return seg, nil
}

// Open existing segment
func OpenSegment(path string, id int) (*Segment, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	seg := &Segment{
		ID:   id,
		Path: path,
		file: file,
	}

	// Read header
	if err := seg.readHeader(); err != nil {
		file.Close()
		return nil, err
	}

	// Get file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	seg.Size = info.Size()

	return seg, nil
}

func (s *Segment) writeHeader() error {
	header := make([]byte, 16)
	binary.LittleEndian.PutUint32(header[0:4], SegmentMagic)
	binary.LittleEndian.PutUint32(header[4:8], SegmentVersion)
	binary.LittleEndian.PutUint32(header[8:12], 0) // DocCount
	binary.LittleEndian.PutUint32(header[12:16], 0) // Reserved

	_, err := s.file.Write(header)
	return err
}

func (s *Segment) readHeader() error {
	header := make([]byte, 16)
	if _, err := s.file.Read(header); err != nil {
		return err
	}

	magic := binary.LittleEndian.Uint32(header[0:4])
	if magic != SegmentMagic {
		return errors.New("invalid segment magic")
	}

	version := binary.LittleEndian.Uint32(header[4:8])
	if version != SegmentVersion {
		return errors.New("unsupported segment version")
	}

	s.DocCount = int(binary.LittleEndian.Uint32(header[8:12]))
	return nil
}

// Append record to segment
func (s *Segment) Append(rec SegmentRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Sealed {
		return errors.New("segment is sealed")
	}

	// Seek to end
	offset, err := s.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	rec.Offset = offset

	// Serialize data
	var data []byte
	if rec.Data != nil {
		data, err = json.Marshal(rec.Data)
		if err != nil {
			return err
		}
	}

	// Build record buffer
	buf := make([]byte, 0, 1024)

	// Type
	buf = append(buf, byte(rec.Type))

	// DocID length + DocID
	docIDBytes := []byte(rec.DocID)
	docIDLen := make([]byte, 2)
	binary.LittleEndian.PutUint16(docIDLen, uint16(len(docIDBytes)))
	buf = append(buf, docIDLen...)
	buf = append(buf, docIDBytes...)

	// Data length + Data
	dataLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataLen, uint32(len(data)))
	buf = append(buf, dataLen...)
	buf = append(buf, data...)

	// CRC32
	crcVal := crc32.ChecksumIEEE(buf)
	crcBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcBytes, crcVal)
	buf = append(buf, crcBytes...)

	// Write total length first
	totalLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(totalLen, uint32(len(buf)))
	if _, err := s.file.Write(totalLen); err != nil {
		return err
	}

	// Write record
	if _, err := s.file.Write(buf); err != nil {
		return err
	}

	// Sync to disk
	if err := s.file.Sync(); err != nil {
		return err
	}

	s.DocCount++
	s.Size = offset + int64(len(totalLen)) + int64(len(buf))

	// Update header
	return s.updateHeaderDocCount()
}

func (s *Segment) updateHeaderDocCount() error {
	// Seek to DocCount position (8 bytes offset)
	if _, err := s.file.Seek(8, io.SeekStart); err != nil {
		return err
	}
	docCountBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(docCountBytes, uint32(s.DocCount))
	_, err := s.file.Write(docCountBytes)
	return err
}

// Read all records from segment
func (s *Segment) ReadAll() ([]SegmentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Seek past header
	if _, err := s.file.Seek(16, io.SeekStart); err != nil {
		return nil, err
	}

	records := make([]SegmentRecord, 0, s.DocCount)

	for {
		rec, err := s.readRecord()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip corrupted record
			continue
		}
		records = append(records, rec)
	}

	return records, nil
}

func (s *Segment) readRecord() (SegmentRecord, error) {
	var rec SegmentRecord

	// Read total length
	totalLenBytes := make([]byte, 4)
	if _, err := s.file.Read(totalLenBytes); err != nil {
		return rec, err
	}
	totalLen := binary.LittleEndian.Uint32(totalLenBytes)

	// Read record data
	buf := make([]byte, totalLen)
	if _, err := io.ReadFull(s.file, buf); err != nil {
		return rec, err
	}

	// Parse record
	offset := 0

	// Type
	if offset >= len(buf) {
		return rec, io.EOF
	}
	rec.Type = RecordType(buf[offset])
	offset++

	// DocID length
	if offset+2 > len(buf) {
		return rec, io.EOF
	}
	docIDLen := binary.LittleEndian.Uint16(buf[offset : offset+2])
	offset += 2

	// DocID
	if offset+int(docIDLen) > len(buf) {
		return rec, io.EOF
	}
	rec.DocID = string(buf[offset : offset+int(docIDLen)])
	offset += int(docIDLen)

	// Data length
	if offset+4 > len(buf) {
		return rec, io.EOF
	}
	dataLen := binary.LittleEndian.Uint32(buf[offset : offset+4])
	offset += 4

	// Data
	if dataLen > 0 {
		if offset+int(dataLen) > len(buf) {
			return rec, io.EOF
		}
		var doc types.Document
		if err := json.Unmarshal(buf[offset:offset+int(dataLen)], &doc); err != nil {
			return rec, err
		}
		rec.Data = doc
		offset += int(dataLen)
	}

	// CRC (verify)
	if offset+4 > len(buf) {
		return rec, io.EOF
	}
	crcStored := binary.LittleEndian.Uint32(buf[offset:])
	crcCalc := crc32.ChecksumIEEE(buf[:offset])
	if crcStored != crcCalc {
		return rec, errors.New("CRC mismatch - corrupted record")
	}

	return rec, nil
}

// Seal segment (mark as read-only)
func (s *Segment) Seal() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Sealed = true
	return s.file.Sync()
}

// Close segment
func (s *Segment) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		return s.file.Close()
	}
	return nil
}