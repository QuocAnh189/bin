package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aq189/bin/internal/bootstrap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Bootstrap the application
	app, err := bootstrap.NewApplication(ctx)
	if err != nil {
		log.Fatalf("failed to bootstrap application: %v", err)
	}

	// Start the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := app.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("received shutdown signal, gracefully stopping...")
	case err := <-errChan:
		log.Printf("server error: %v", err)
	}

	// Graceful shutdown
	if err := app.Stop(ctx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}

	log.Println("root server stopped")
}
