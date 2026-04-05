package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/martin/primordia/internal/server"
	"github.com/martin/primordia/internal/world"
)

const (
	TickRate      = 30 * time.Millisecond
	BroadcastRate = 40 * time.Millisecond
)

// main bootstraps process state, starts simulation/broadcast loops, and serves HTTP endpoints.
func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w := world.New()
	srv := server.New(w, BroadcastRate)
	var tickIntervalNS atomic.Int64
	var running atomic.Bool
	running.Store(true)
	tickIntervalNS.Store(TickRate.Nanoseconds())
	srv.SetSpeedController(1, func(rate float64) {
		if rate <= 0 {
			return
		}
		tickIntervalNS.Store(time.Duration(float64(TickRate) / rate).Nanoseconds())
	})
	srv.SetControlHandler(func(action string) bool {
		switch action {
		case "start":
			running.Store(true)
			return true
		case "stop":
			running.Store(false)
			return true
		case "restart":
			w.Reset()
			running.Store(true)
			return true
		default:
			return false
		}
	})

	engineDone := make(chan struct{})
	go func() {
		for {
			tickDelay := time.Duration(tickIntervalNS.Load())
			timer := time.NewTimer(tickDelay)
			select {
			case <-timer.C:
				if running.Load() {
					w.Tick()
					if w.OrganismCount() == 0 {
						running.Store(false)
					}
				}
			case <-engineDone:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			}
		}
	}()

	go srv.BroadcastLoop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.WSHandler)
	mux.HandleFunc("/speed", srv.SpeedHandler)
	mux.HandleFunc("/control", srv.ControlHandler)
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
