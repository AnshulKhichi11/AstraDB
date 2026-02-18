package engine

import "time"

const (
	DefaultDataDir      = "./data"
	DefaultDBsDir       = "./data/databases"
	DefaultWALDir       = "./data/wal"
	DefaultWALFile      = "./data/wal/wal.log"
	DefaultWALArchiveDir = "./data/wal/archive"
	DefaultMaxWALBytes  = 5 * 1024 * 1024
	DefaultMaxDocBytes  = 16 * 1024 * 1024

	DefaultCompactionInterval    = 5 * 60
	DefaultCompactionThreshold   = 3
	DefaultDeadSpaceThreshold    = 30

	// NEW: WAL settings
	DefaultCheckpointInterval    = 60        // 60 seconds
	DefaultCheckpointWALSize     = 10 * 1024 * 1024 // 10MB
	DefaultWALSyncMode          = "batch"   // immediate/batch/async
	DefaultWALBatchSize         = 100       // entries before fsync
	DefaultWALBatchTimeout      = 1         // seconds
)

type WALSyncMode string

const (
	WALSyncImmediate WALSyncMode = "immediate" // fsync every write (safest, slowest)
	WALSyncBatch     WALSyncMode = "batch"     // fsync every N writes (balanced)
	WALSyncAsync     WALSyncMode = "async"     // fsync on timer (fastest, riskiest)
)

type Config struct {
	DataDir     string
	DBsDir      string
	WALDir      string
	WALFile     string
	WALArchiveDir string
	MaxWALBytes int64
	MaxDocBytes int

	CompactionInterval   int
	CompactionThreshold  int
	DeadSpaceThreshold   int
	EnableAutoCompaction bool

	// NEW: WAL configuration
	CheckpointInterval time.Duration
	CheckpointWALSize  int64
	WALSyncMode       WALSyncMode
	WALBatchSize      int
	WALBatchTimeout   time.Duration
	EnableWALArchive  bool
}

func DefaultConfig() Config {
	return Config{
		DataDir:     DefaultDataDir,
		DBsDir:      DefaultDBsDir,
		WALDir:      DefaultWALDir,
		WALFile:     DefaultWALFile,
		WALArchiveDir: DefaultWALArchiveDir,
		MaxWALBytes: DefaultMaxWALBytes,
		MaxDocBytes: DefaultMaxDocBytes,

		CompactionInterval:   DefaultCompactionInterval,
		CompactionThreshold:  DefaultCompactionThreshold,
		DeadSpaceThreshold:   DefaultDeadSpaceThreshold,
		EnableAutoCompaction: true,

		CheckpointInterval: time.Duration(DefaultCheckpointInterval) * time.Second,
		CheckpointWALSize:  DefaultCheckpointWALSize,
		WALSyncMode:       WALSyncBatch,
		WALBatchSize:      DefaultWALBatchSize,
		WALBatchTimeout:   time.Duration(DefaultWALBatchTimeout) * time.Second,
		EnableWALArchive:  true,
	}
}