package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sentinel-proxy/api/internal/auth"
	"github.com/sentinel-proxy/api/internal/config"
	"github.com/sentinel-proxy/api/internal/handlers"
	"github.com/sentinel-proxy/api/internal/storage"
)

func main() {
	cfg := config.Load()

	// Initialize storage
	stor, err := storage.New(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize handlers
	logsHandler := handlers.NewLogsHandler(stor)
	incidentsHandler := handlers.NewIncidentsHandler(stor)
	rulesHandler := handlers.NewRulesHandler(stor)
	actionsHandler, err := handlers.NewActionsHandler(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize actions handler: %v", err)
	}
	aiHandler := handlers.NewAIHandler(stor)

	// Admin auth middleware
	adminAuth := auth.AdminTokenMiddleware(cfg.AdminToken)

	// Setup routes
	mux := http.NewServeMux()

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Sentinel-Admin-Token")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}

	// Health check (no auth)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Protected routes (with CORS)
	mux.Handle("/logs", corsMiddleware(adminAuth(http.HandlerFunc(logsHandler.GetLogs))))
	mux.Handle("/incidents", corsMiddleware(adminAuth(http.HandlerFunc(incidentsHandler.GetIncidents))))
	mux.Handle("/incidents/", corsMiddleware(adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/close") {
			incidentsHandler.CloseIncident(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))))
	mux.Handle("/rules", corsMiddleware(adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rulesHandler.GetRules(w, r)
		case http.MethodPost:
			rulesHandler.CreateRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))
	mux.Handle("/rules/", corsMiddleware(adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			rulesHandler.UpdateRule(w, r)
		case http.MethodDelete:
			rulesHandler.DeleteRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))
	mux.Handle("/block", corsMiddleware(adminAuth(http.HandlerFunc(actionsHandler.Block))))
	mux.Handle("/unblock", corsMiddleware(adminAuth(http.HandlerFunc(actionsHandler.Unblock))))
	mux.Handle("/redirect", corsMiddleware(adminAuth(http.HandlerFunc(actionsHandler.Redirect))))
	mux.Handle("/unredirect", corsMiddleware(adminAuth(http.HandlerFunc(actionsHandler.Unredirect))))
	mux.Handle("/ai/analyze", corsMiddleware(adminAuth(http.HandlerFunc(aiHandler.Analyze))))

	// Start server
	addr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("Sentinel API listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
