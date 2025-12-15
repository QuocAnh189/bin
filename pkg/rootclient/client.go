package rootclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the Root Server client SDK
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Config holds client configuration
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// New creates a new Root Server client
func New(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Auth returns the authentication service client
func (c *Client) Auth() *AuthClient {
	return &AuthClient{client: c}
}

// Session returns the session service client
func (c *Client) Session() *SessionClient {
	return &SessionClient{client: c}
}

// Registry returns the registry service client
func (c *Client) Registry() *RegistryClient {
	return &RegistryClient{client: c}
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// AuthClient handles authentication operations
type AuthClient struct {
	client *Client
}

// IssueTokenRequest represents a token issuance request
type IssueTokenRequest struct {
	Subject  string         `json:"subject"`
	Roles    []string       `json:"roles,omitempty"`
	Audience string         `json:"audience,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	Token     string    `json:"token"`
	Type      string    `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

// IssueToken requests a new JWT token
func (a *AuthClient) IssueToken(ctx context.Context, req IssueTokenRequest) (*TokenResponse, error) {
	var resp TokenResponse
	if err := a.client.doRequest(ctx, http.MethodPost, "/auth/token", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ValidateToken validates a JWT token
func (a *AuthClient) ValidateToken(ctx context.Context, token string) error {
	req := map[string]string{"token": token}
	return a.client.doRequest(ctx, http.MethodPost, "/auth/validate", req, nil)
}

// SessionClient handles session operations
type SessionClient struct {
	client *Client
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	UserID    string         `json:"user_id"`
	ServiceID string         `json:"service_id"`
	Data      map[string]any `json:"data"`
	TTL       int            `json:"ttl"` // minutes
}

// Session represents a session
type Session struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	ServiceID string         `json:"service_id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Create creates a new session
func (s *SessionClient) Create(ctx context.Context, req CreateSessionRequest) (*Session, error) {
	var session Session
	if err := s.client.doRequest(ctx, http.MethodPost, "/session", req, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// Get retrieves a session by ID
func (s *SessionClient) Get(ctx context.Context, id string) (*Session, error) {
	var session Session
	if err := s.client.doRequest(ctx, http.MethodGet, "/session/"+id, nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// Update updates a session
func (s *SessionClient) Update(ctx context.Context, id string, data map[string]any) error {
	req := map[string]any{"data": data}
	return s.client.doRequest(ctx, http.MethodPut, "/session/"+id, req, nil)
}

// Delete deletes a session
func (s *SessionClient) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, http.MethodDelete, "/session/"+id, nil, nil)
}

// RegistryClient handles service registry operations
type RegistryClient struct {
	client *Client
}

// RegisterRequest represents a service registration request
type RegisterRequest struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Endpoints      []string          `json:"endpoints"`
	Capabilities   []string          `json:"capabilities"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	HealthCheckURL string            `json:"health_check_url,omitempty"`
}

// Service represents a registered service
type Service struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Endpoints      []string          `json:"endpoints"`
	Capabilities   []string          `json:"capabilities"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Status         string            `json:"status"`
	RegisteredAt   time.Time         `json:"registered_at"`
	LastHeartbeat  time.Time         `json:"last_heartbeat"`
	HealthCheckURL string            `json:"health_check_url,omitempty"`
}

// Register registers a service with the root server
func (r *RegistryClient) Register(ctx context.Context, req RegisterRequest) (*Service, error) {
	var service Service
	if err := r.client.doRequest(ctx, http.MethodPost, "/registry/register", req, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

// Deregister removes a service from the registry
func (r *RegistryClient) Deregister(ctx context.Context, id string) error {
	return r.client.doRequest(ctx, http.MethodDelete, "/registry/deregister/"+id, nil, nil)
}

// Discover finds services by capability
func (r *RegistryClient) Discover(ctx context.Context, capability string) ([]*Service, error) {
	var services []*Service
	path := "/registry/discover"
	if capability != "" {
		path += "?capability=" + capability
	}
	if err := r.client.doRequest(ctx, http.MethodGet, path, nil, &services); err != nil {
		return nil, err
	}
	return services, nil
}

// Heartbeat sends a heartbeat for a service
func (r *RegistryClient) Heartbeat(ctx context.Context, id string) error {
	return r.client.doRequest(ctx, http.MethodPut, "/registry/heartbeat/"+id, nil, nil)
}
