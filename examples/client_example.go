package main

import (
	"context"
	"log"
	"time"

	"root/pkg/rootclient"
)

func main() {
	// Initialize Root Server client
	client := rootclient.New(rootclient.Config{
		BaseURL: "http://localhost:8080",
		APIKey:  "development-api-key",
		Timeout: 10 * time.Second,
	})

	ctx := context.Background()

	// Example 1: Register your service
	log.Println("Registering service...")
	service, err := client.Registry().Register(ctx, rootclient.RegisterRequest{
		ID:           "example-svc-1",
		Name:         "example-service",
		Version:      "1.0.0",
		Endpoints:    []string{"http://localhost:9090"},
		Capabilities: []string{"example", "demo"},
		Metadata: map[string]string{
			"environment": "development",
		},
		HealthCheckURL: "http://localhost:9090/health",
	})
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	log.Printf("Service registered: %s (ID: %s)\n", service.Name, service.ID)

	// Example 2: Issue a token
	log.Println("Issuing token...")
	token, err := client.Auth().IssueToken(ctx, rootclient.IssueTokenRequest{
		Subject: "user-123",
		Roles:   []string{"admin", "user"},
		Metadata: map[string]any{
			"service": "example-service",
		},
	})
	if err != nil {
		log.Fatalf("Failed to issue token: %v", err)
	}
	log.Printf("Token issued: %s... (expires: %s)\n", token.Token[:20], token.ExpiresAt)

	// Example 3: Create a session
	log.Println("Creating session...")
	session, err := client.Session().Create(ctx, rootclient.CreateSessionRequest{
		UserID:    "user-123",
		ServiceID: "example-svc-1",
		Data: map[string]any{
			"theme":    "dark",
			"language": "en",
		},
		TTL: 60, // 60 minutes
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	log.Printf("Session created: %s (expires: %s)\n", session.ID, session.ExpiresAt)

	// Example 4: Discover services
	log.Println("Discovering services...")
	services, err := client.Registry().Discover(ctx, "example")
	if err != nil {
		log.Fatalf("Failed to discover services: %v", err)
	}
	log.Printf("Found %d services with 'example' capability\n", len(services))
	for _, svc := range services {
		log.Printf("  - %s v%s (%s)\n", svc.Name, svc.Version, svc.Status)
	}

	// Example 5: Send heartbeat
	log.Println("Sending heartbeat...")
	if err := client.Registry().Heartbeat(ctx, service.ID); err != nil {
		log.Fatalf("Failed to send heartbeat: %v", err)
	}
	log.Println("Heartbeat sent successfully")

	// Example 6: Retrieve session
	log.Println("Retrieving session...")
	retrievedSession, err := client.Session().Get(ctx, session.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve session: %v", err)
	}
	log.Printf("Session data: %+v\n", retrievedSession.Data)

	// Cleanup
	log.Println("Cleaning up...")
	if err := client.Session().Delete(ctx, session.ID); err != nil {
		log.Printf("Warning: Failed to delete session: %v", err)
	}
	if err := client.Registry().Deregister(ctx, service.ID); err != nil {
		log.Printf("Warning: Failed to deregister service: %v", err)
	}

	log.Println("Example completed successfully!")
}
