package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ak1m1tsu/lrclib/internal/handler"
	"github.com/ak1m1tsu/lrclib/internal/lrclib"
	"github.com/gorilla/mux"
)

//go:embed static
var staticFiles embed.FS

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	client := lrclib.New()
	tmpl := handler.Templates{}

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HomeHandler(tmpl)).Methods(http.MethodGet)
	r.HandleFunc("/search", handler.SearchHandler(client, tmpl)).Methods(http.MethodGet)
	r.HandleFunc("/lyrics/{id}", handler.LyricsHandler(client, tmpl)).Methods(http.MethodGet)
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles))).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler.SecureHeaders(handler.Logger(handler.Recover(r))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on %s", *addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
