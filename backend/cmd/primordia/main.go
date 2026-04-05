package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/martin/primordia/internal/server"
	"github.com/martin/primordia/internal/world"
)

const (
	TickRate      = 30 * time.Millisecond
	BroadcastRate = 40 * time.Millisecond
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w := world.New()
	srv := server.New(w, BroadcastRate)

	engineDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(TickRate)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.Tick()
			case <-engineDone:
				return
			}
		}
	}()

	go srv.BroadcastLoop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.WSHandler)
	httpServer := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		<-ctx.Done()
		close(engineDone)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	log.Println("Primordia engine running on :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("critical server failure: %v", err)
	}
}
