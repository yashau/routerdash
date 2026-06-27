package routerdash

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	collector *Collector
	static    fs.FS
	mux       *http.ServeMux
	version   string
}

func NewServer(collector *Collector, static fs.FS, version string) http.Handler {
	s := &Server{collector: collector, static: static, mux: http.NewServeMux(), version: version}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	s.mux.HandleFunc("GET /api/version", s.json(func(_ *http.Request) any {
		return VersionInfo{Version: s.version}
	}))
	s.mux.HandleFunc("GET /api/summary", s.json(func(r *http.Request) any {
		return s.collector.Summary(r.Context())
	}))
	s.mux.HandleFunc("GET /api/metrics", s.json(func(r *http.Request) any {
		return s.collector.Metrics(r.Context())
	}))
	s.mux.HandleFunc("GET /api/tailscale", s.json(func(r *http.Request) any {
		return s.collector.TailscalePage(r.Context(), pageRequest(r, 10))
	}))
	s.mux.HandleFunc("GET /api/rathole", s.json(func(r *http.Request) any {
		return s.collector.Rathole(r.Context())
	}))
	s.mux.HandleFunc("GET /api/firewall", s.json(func(r *http.Request) any {
		return s.collector.Firewall(r.Context())
	}))
	s.mux.HandleFunc("GET /api/routes", s.json(func(r *http.Request) any {
		return s.collector.RoutesPage(r.Context(), pageRequest(r, 50))
	}))
	s.mux.HandleFunc("GET /api/dhcp", s.json(func(r *http.Request) any {
		return s.collector.DHCPLeasesPage(r.Context(), pageRequest(r, 50))
	}))
	s.mux.HandleFunc("GET /api/frr", s.json(func(r *http.Request) any {
		return s.collector.FRR(r.Context())
	}))
	s.mux.HandleFunc("POST /api/diagnostics", s.diagnostic)
	s.mux.HandleFunc("/api/", http.NotFound)
	s.mux.HandleFunc("/", s.spa)
}

func (s *Server) json(fn func(*http.Request) any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := contextWithTimeout(r, 12*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		writeJSON(w, fn(r))
	}
}

func (s *Server) diagnostic(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 45*time.Second)
	defer cancel()
	var req DiagnosticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	writeJSON(w, s.collector.Diagnostic(ctx, req))
}

func (s *Server) spa(w http.ResponseWriter, r *http.Request) {
	filePath := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if filePath == "." {
		filePath = "index.html"
	}
	if filePath == "index.html" {
		serveAppShell(w, r, s.static)
		return
	}
	if !fs.ValidPath(filePath) {
		serveAppShell(w, r, s.static)
		return
	}
	info, err := fs.Stat(s.static, filePath)
	if err != nil || info.IsDir() {
		serveAppShell(w, r, s.static)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFileFS(w, r, s.static, filePath)
}

func serveAppShell(w http.ResponseWriter, r *http.Request, static fs.FS) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Clear-Site-Data", `"cache"`)
	http.ServeFileFS(w, r, static, "index.html")
}

func contextWithTimeout(r *http.Request, timeout time.Duration) (context.Context, func()) {
	return context.WithTimeout(r.Context(), timeout)
}

func pageRequest(r *http.Request, defaultPageSize int) PageRequest {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	return PageRequest{Page: page, PageSize: pageSize}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil && !errors.Is(err, http.ErrAbortHandler) {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}
