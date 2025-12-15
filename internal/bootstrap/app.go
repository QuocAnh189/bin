package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/aq189/bin/internal/domain/config"
	"github.com/aq189/bin/internal/handler/auth"
	"github.com/aq189/bin/internal/handler/health"
	"github.com/aq189/bin/internal/handler/registry"
	sessionhandler "github.com/aq189/bin/internal/handler/session"
	"github.com/aq189/bin/internal/middleware"
	"github.com/aq189/bin/internal/repository/memory"
	"github.com/aq189/bin/internal/repository/postgres"
	"github.com/aq189/bin/internal/repository/redis"
	"github.com/aq189/bin/internal/server"
	authsvc "github.com/aq189/bin/internal/service/auth"
	registrysvc "github.com/aq189/bin/internal/service/registry"
	sessionsvc "github.com/aq189/bin/internal/service/session"
	"github.com/aq189/bin/pkg/jwt"
	"github.com/aq189/bin/pkg/logger"
)

// Application represents the root server application lifecycle
type Application struct {
	config *config.Config
	server *server.Server
	logger logger.Logger

	// Repositories
	sessionRepo  sessionsvc.SessionRepository
	registryRepo registrysvc.RegistryRepository
	configRepo   config.ConfigRepository

	// Services
	authService     *authsvc.Service
	sessionService  *sessionsvc.Service
	registryService *registrysvc.Service

	// Cleanup functions
	cleanup []func() error
}

// NewApplication initializes and wires up all dependencies
func NewApplication(ctx context.Context) (*Application, error) {
	app := &Application{
		cleanup: make([]func() error, 0),
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	app.config = cfg

	// Initialize logger
	app.logger = logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	// Initialize repositories based on configuration
	if err := app.initRepositories(ctx); err != nil {
		return nil, fmt.Errorf("init repositories: %w", err)
	}

	// Initialize services
	if err := app.initServices(); err != nil {
		return nil, fmt.Errorf("init services: %w", err)
	}

	// Initialize HTTP server
	if err := app.initServer(); err != nil {
		return nil, fmt.Errorf("init server: %w", err)
	}

	return app, nil
}

// initRepositories sets up data persistence layers
func (app *Application) initRepositories(ctx context.Context) error {
	switch app.config.Storage.Type {
	case "redis":
		redisRepo, err := redis.NewRepository(ctx, redis.Config{
			Addr:     app.config.Storage.Redis.Addr,
			Password: app.config.Storage.Redis.Password,
			DB:       app.config.Storage.Redis.DB,
		})
		if err != nil {
			return fmt.Errorf("redis repository: %w", err)
		}
		app.sessionRepo = redisRepo
		app.cleanup = append(app.cleanup, redisRepo.Close)

	case "postgres":
		pgRepo, err := postgres.NewRepository(ctx, postgres.Config{
			Host:     app.config.Storage.Postgres.Host,
			Port:     app.config.Storage.Postgres.Port,
			User:     app.config.Storage.Postgres.User,
			Password: app.config.Storage.Postgres.Password,
			Database: app.config.Storage.Postgres.Database,
		})
		if err != nil {
			return fmt.Errorf("postgres repository: %w", err)
		}
		app.registryRepo = pgRepo
		app.cleanup = append(app.cleanup, pgRepo.Close)

	default:
		// Use in-memory for development
		app.sessionRepo = memory.NewSessionRepository()
		app.registryRepo = memory.NewRegistryRepository()
		app.configRepo = memory.NewConfigRepository()
	}

	return nil
}

// initServices initializes business logic services
func (app *Application) initServices() error {
	// JWT service for token operations
	jwtService, err := jwt.NewService(jwt.Config{
		Secret:          app.config.JWT.Secret,
		AccessTokenTTL:  time.Duration(app.config.JWT.AccessTokenTTL) * time.Minute,
		RefreshTokenTTL: time.Duration(app.config.JWT.RefreshTokenTTL) * time.Hour,
		Issuer:          "root-server",
	})
	if err != nil {
		return fmt.Errorf("jwt service: %w", err)
	}

	// Auth service
	app.authService = authsvc.New(authsvc.Config{
		JWTService: jwtService,
		Logger:     app.logger,
	})

	// Session service
	app.sessionService = sessionsvc.New(sessionsvc.Config{
		Repository:    app.sessionRepo,
		DefaultTTL:    time.Duration(app.config.Session.DefaultTTL) * time.Minute,
		CleanupPeriod: time.Duration(app.config.Session.CleanupPeriod) * time.Minute,
		Logger:        app.logger,
	})

	// Registry service
	app.registryService = registrysvc.New(registrysvc.Config{
		Repository:          app.registryRepo,
		HealthCheckInterval: time.Duration(app.config.Registry.HealthCheckInterval) * time.Second,
		HealthCheckTimeout:  time.Duration(app.config.Registry.HealthCheckTimeout) * time.Second,
		Logger:              app.logger,
	})

	return nil
}

// initServer sets up the HTTP server with routes and middleware
func (app *Application) initServer() error {
	// Initialize handlers
	authHandler := auth.NewHandler(app.authService, app.logger)
	sessionHandler := sessionhandler.NewHandler(app.sessionService, app.logger)
	registryHandler := registry.NewHandler(app.registryService, app.logger)
	healthHandler := health.NewHandler(app.logger)

	// Build middleware chain
	middlewares := []server.Middleware{
		middleware.RequestID(),
		middleware.Logger(app.logger),
		middleware.Recovery(app.logger),
		middleware.CORS(app.config.Server.CORS),
	}

	// Protected routes middleware (requires authentication)
	authMiddleware := middleware.Authenticate(app.authService)

	// Create server with configuration
	srv, err := server.New(server.Config{
		Addr:         app.config.Server.Addr,
		ReadTimeout:  time.Duration(app.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(app.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(app.config.Server.IdleTimeout) * time.Second,
		TLS: server.TLSConfig{
			Enabled:  app.config.Server.TLS.Enabled,
			CertFile: app.config.Server.TLS.CertFile,
			KeyFile:  app.config.Server.TLS.KeyFile,
		},
		Middlewares: middlewares,
	})
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	// Register public routes
	srv.GET("/health", healthHandler.Health)
	srv.GET("/ready", healthHandler.Ready)

	// Register protected routes
	srv.POST("/auth/token", authHandler.IssueToken, authMiddleware)
	srv.POST("/auth/validate", authHandler.ValidateToken, authMiddleware)
	srv.POST("/auth/refresh", authHandler.RefreshToken, authMiddleware)
	srv.POST("/auth/revoke", authHandler.RevokeToken, authMiddleware)

	srv.POST("/session", sessionHandler.Create, authMiddleware)
	srv.GET("/session/:id", sessionHandler.Get, authMiddleware)
	srv.PUT("/session/:id", sessionHandler.Update, authMiddleware)
	srv.DELETE("/session/:id", sessionHandler.Delete, authMiddleware)

	srv.POST("/registry/register", registryHandler.Register, authMiddleware)
	srv.DELETE("/registry/deregister/:id", registryHandler.Deregister, authMiddleware)
	srv.GET("/registry/services", registryHandler.ListServices, authMiddleware)
	srv.GET("/registry/discover", registryHandler.Discover, authMiddleware)
	srv.PUT("/registry/heartbeat/:id", registryHandler.Heartbeat, authMiddleware)

	app.server = srv
	return nil
}

// Start begins the application lifecycle
func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("starting root server", map[string]any{
		"addr": app.config.Server.Addr,
		"tls":  app.config.Server.TLS.Enabled,
	})

	// Start background services
	go app.sessionService.StartCleanup(ctx)
	go app.registryService.StartHealthChecks(ctx)

	// Start HTTP server
	return app.server.Start()
}

// Stop gracefully shuts down the application
func (app *Application) Stop(ctx context.Context) error {
	app.logger.Info("stopping root server", map[string]any{})

	// Shutdown HTTP server
	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("server shutdown error", map[string]any{"error": err})
	}

	// Run cleanup functions in reverse order
	for i := len(app.cleanup) - 1; i >= 0; i-- {
		if err := app.cleanup[i](); err != nil {
			app.logger.Error("cleanup error", map[string]any{"error": err})
		}
	}

	app.logger.Info("root server stopped successfully", map[string]any{})
	return nil
}
