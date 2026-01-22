package frontend

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MustafaMertSandal/SNMP_Task/internal/config"
	"github.com/MustafaMertSandal/SNMP_Task/internal/db"
)

//go:embed web/*
var webFS embed.FS

func StartWebServer(ctx context.Context, cfg *config.Config, store *db.Store) (*http.Server, error) {
	if cfg.Web.Enabled != nil && !*cfg.Web.Enabled {
		return nil, nil
	}

	mux := http.NewServeMux()

	// Static UI
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, fmt.Errorf("embed web fs: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API
	mux.HandleFunc("/api/devices", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/devices" {
			http.NotFound(writer, request)
			return
		}
		if store == nil {
			http.Error(writer, "Database is disabled", http.StatusServiceUnavailable)
			return
		}
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", http.MethodGet)
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		devices, err := store.ListDevices(request.Context())
		if err != nil {
			http.Error(writer, fmt.Sprintf("list devices: %v", err), http.StatusInternalServerError)
			return
		}
		writeJSON(writer, devices)
	})

	mux.HandleFunc("/api/devices/", func(writer http.ResponseWriter, request *http.Request) {
		// /api/devices/{id}/(snmp|interfaces|routes)
		if store == nil {
			http.Error(writer, "database is disabled", http.StatusServiceUnavailable)
			return
		}
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", http.MethodGet)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(request.URL.Path, "/api/devices/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) < 2 {
			http.NotFound(writer, request)
			return
		}

		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || id <= 0 {
			http.Error(writer, "invalid device id", http.StatusBadRequest)
			return
		}

		endpoint := parts[1]

		switch endpoint {
		case "snmp":
			rows, err := store.GetRouterSNMP1mAllByDeviceID(request.Context(), id)
			if err != nil {
				http.Error(writer, fmt.Sprintf("snmp query: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSON(writer, rows)
		case "interfaces":
			rows, err := store.GetRouterInterfaceMetrics1mAllByDeviceID(request.Context(), id)
			if err != nil {
				http.Error(writer, fmt.Sprintf("interfaces query: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSON(writer, rows)
		case "routes":
			rows, err := store.GetRouterIPRoutes1mAllByDeviceID(request.Context(), id)
			if err != nil {
				http.Error(writer, fmt.Sprintf("routes query: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSON(writer, rows)
		default:
			http.NotFound(writer, request)
		}
	})

	srv := &http.Server{
		Addr:              cfg.Web.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctxShutdown)
	}()

	go func() {
		_ = srv.ListenAndServe()
	}()

	return srv, nil
}

func writeJSON(writer http.ResponseWriter, v any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
