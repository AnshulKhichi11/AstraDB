package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"testDB/internal/api"
	"testDB/internal/engine"
)

func main() {
	// Command line flags
	walMode := flag.String("wal-mode", "batch", "WAL sync mode: immediate, batch, async")
	walBatch := flag.Int("wal-batch", 100, "WAL batch size")
	checkpointInterval := flag.Int("checkpoint", 60, "Checkpoint interval (seconds)")
	
	flag.Parse()

	cfg := engine.DefaultConfig()
	
	// Apply flags
	cfg.WALSyncMode = engine.WALSyncMode(*walMode)
	cfg.WALBatchSize = *walBatch
	cfg.CheckpointInterval = time.Duration(*checkpointInterval) * time.Second

	eng, err := engine.New(cfg)
	if err != nil {
		panic(err)
	}

	// Start auto-compaction
	eng.StartAutoCompaction(5 * time.Minute)

	router := api.NewRouter(eng)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n🛑 Shutdown signal received...")
		eng.Shutdown()
		os.Exit(0)
	}()

	fmt.Println("✅ AstraDB Server: http://localhost:8080")
	fmt.Println("📁 Storage:", cfg.DBsDir+"/<db>/collections/<collection>/segments/")
	fmt.Println("🪵 WAL Mode:", cfg.WALSyncMode)
	fmt.Println("🔄 Auto-compaction: Every 5 minutes")
	fmt.Println("💾 Checkpoint: Every", cfg.CheckpointInterval)

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}