package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/maelstrom/v3/pkg/kernel"
	"github.com/maelstrom/v3/pkg/statechart"
)

func main() {
	log.Println("[main] Starting Maelstrom")

	// Create cancellation context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("[main] Received signal: %v", sig)
		cancel()
	}()

	// Create the statechart engine (shared across all runtimes)
	engine := statechart.NewEngine()
	log.Println("[main] Statechart engine created")

	// Create the kernel (orchestrates bootstrap)
	k := kernel.NewWithEngine(engine)
	log.Println("[main] Kernel created")

	// Start the kernel (runs bootstrap sequence)
	if err := k.Start(ctx); err != nil {
		if err == context.Canceled {
			log.Println("[main] Shutdown requested")
		} else {
			log.Printf("[main] Kernel error: %v", err)
			os.Exit(1)
		}
	}

	log.Println("[main] Maelstrom shutdown complete")
}
