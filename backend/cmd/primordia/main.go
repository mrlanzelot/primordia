package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/ws", srv.WSHandler)
	mux.HandleFunc("/speed", srv.SpeedHandler)
	mux.HandleFunc("/control", srv.ControlHandler)
	mux.HandleFunc("/api/ws", srv.WSHandler)
	mux.HandleFunc("/api/speed", srv.SpeedHandler)
	mux.HandleFunc("/api/control", srv.ControlHandler)
	registerStaticRoutes(mux)
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

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func registerStaticRoutes(mux *http.ServeMux) {
	webRoot, ok := resolveWebRoot()
	if !ok {
		log.Printf("frontend assets missing; static routes disabled")
		return
	}

	assetFS := http.FileServer(http.Dir(webRoot))
	mux.Handle("/assets/", assetFS)
	mux.Handle("/favicon.ico", assetFS)
	mux.Handle("/vite.svg", assetFS)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cleanPath := filepath.Clean(r.URL.Path)
		if strings.HasPrefix(cleanPath, "/api/") || cleanPath == "/api" {
			http.NotFound(w, r)
			return
		}

		if cleanPath == "/" {
			http.ServeFile(w, r, filepath.Join(webRoot, "index.html"))
			return
		}

		target := filepath.Join(webRoot, strings.TrimPrefix(cleanPath, "/"))
		info, err := os.Stat(target)
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			http.Error(w, "failed to load asset", http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, filepath.Join(webRoot, "index.html"))
	})

	log.Printf("serving frontend assets from %s", webRoot)
}

func resolveWebRoot() (string, bool) {
	paths := []string{
		"/app/web",
		"./web",
		"../frontend/dist",
	}
	for _, p := range paths {
		index := filepath.Join(p, "index.html")
		if _, err := os.Stat(index); err == nil {
			return p, true
		}
	}
	return "", false
}
