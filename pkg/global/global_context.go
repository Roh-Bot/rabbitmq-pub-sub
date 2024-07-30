package global

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
)

func initialize() {
	ctx, cancel = context.WithCancel(context.Background())

	// Set up signal handling to cancel context on SIGINT and SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s, initiating shutdown...", sig)
		cancel()
		log.Println("Shutdown signal braodcasted")
	}()
}

// CancellationContext returns the global context
func CancellationContext() context.Context {
	once.Do(initialize)
	return ctx
}
