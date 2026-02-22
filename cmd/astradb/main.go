package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"testDB/internal/api"
	"testDB/internal/engine"
)

func main() {
	// -----------------------------
	// Command Line Flags
	// -----------------------------
	walMode := flag.String("wal-mode", "batch", "WAL sync mode: immediate, batch, async")
	walBatch := flag.Int("wal-batch", 100, "WAL batch size")
	checkpointInterval := flag.Int("checkpoint", 60, "Checkpoint interval (seconds)")
	flag.Parse()

	cfg := engine.DefaultConfig()

	cfg.WALSyncMode = engine.WALSyncMode(*walMode)
	cfg.WALBatchSize = *walBatch
	cfg.CheckpointInterval = time.Duration(*checkpointInterval) * time.Second

	// -----------------------------
	// Initialize Engine
	// -----------------------------
	eng, err := engine.New(cfg)
	if err != nil {
		log.Fatalf("Engine initialization failed: %v", err)
	}

	eng.StartAutoCompaction(5 * time.Minute)

	router := api.NewRouter(eng)

	// Detect mode
	authEnabled := os.Getenv("ASTRA_AUTH") == "true"
	mode := "DEVELOPMENT"
	if authEnabled {
		mode = "PRODUCTION"
	}

	// -----------------------------
	// Startup Information
	// -----------------------------
	log.Println("--------------------------------------------------")
	log.Println("AstraDB Server Starting")
	log.Println("--------------------------------------------------")
	log.Printf("Mode               : %s", mode)
	log.Printf("Server Address     : http://localhost:8080")
	log.Printf("Storage Path       : %s/<db>/collections/<collection>/segments/", cfg.DBsDir)
	log.Printf("WAL Mode           : %s", cfg.WALSyncMode)
	log.Printf("WAL Batch Size     : %d", cfg.WALBatchSize)
	log.Printf("Checkpoint Interval: %s", cfg.CheckpointInterval)
	log.Printf("Auto Compaction    : every 5m")
	log.Println("--------------------------------------------------")
	log.Println("Server is ready to accept connections")
	log.Println("")

	// -----------------------------
	// Graceful Shutdown
	// -----------------------------
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutdown signal received")
		eng.Shutdown()
		log.Println("AstraDB stopped")
		os.Exit(0)
	}()

	// -----------------------------
	// Start HTTP Server
	// -----------------------------
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
