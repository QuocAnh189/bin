package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// HandlerFunc is the signature for HTTP handler functions
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Config holds HTTP server configuration
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	TLS          TLSConfig
	Middlewares  []Middleware
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// Server wraps net/http server with routing and middleware
type Server struct {
	config     Config
	httpServer *http.Server
	mux        *http.ServeMux
	middleware []Middleware
}

// New creates a new HTTP server instance
func New(config Config) (*Server, error) {
	mux := http.NewServeMux()

	srv := &Server{
		config: config,
		httpServer: &http.Server{
			Addr:         config.Addr,
			Handler:      mux,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		mux:        mux,
		middleware: config.Middlewares,
	}

	return srv, nil
}

// GET registers a GET route
func (s *Server) GET(pattern string, handler HandlerFunc, middleware ...Middleware) {
	s.handle(http.MethodGet, pattern, handler, middleware...)
}

// POST registers a POST route
func (s *Server) POST(pattern string, handler HandlerFunc, middleware ...Middleware) {
	s.handle(http.MethodPost, pattern, handler, middleware...)
}

// PUT registers a PUT route
func (s *Server) PUT(pattern string, handler HandlerFunc, middleware ...Middleware) {
	s.handle(http.MethodPut, pattern, handler, middleware...)
}

// DELETE registers a DELETE route
func (s *Server) DELETE(pattern string, handler HandlerFunc, middleware ...Middleware) {
	s.handle(http.MethodDelete, pattern, handler, middleware...)
}

// handle registers a route with method-based filtering and middleware
func (s *Server) handle(method, pattern string, handler HandlerFunc, middleware ...Middleware) {
	// Wrap handler with method checking
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	})

	// Apply route-specific middleware
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	// Apply global middleware
	for i := len(s.middleware) - 1; i >= 0; i-- {
		h = s.middleware[i](h)
	}

	s.mux.Handle(pattern, h)
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	if s.config.TLS.Enabled {
		return s.httpServer.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile)
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
