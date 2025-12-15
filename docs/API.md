# Root Server API Documentation

## Overview

The Root Server provides RESTful APIs for authentication, session management, service registry, and configuration management.

All authenticated endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer <token>
```

## Base URL

- Development: `http://localhost:8080`
- Production: `https://root.company.internal`

## Common Headers

| Header | Description | Required |
|--------|-------------|----------|
| Content-Type | application/json | Yes |
| Authorization | Bearer <token> | For protected endpoints |
| X-Request-ID | Request correlation ID | Optional (auto-generated) |

## Response Format

### Success Response

```json
{
  "data": { ... }
}
```

### Error Response

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "request_id": "abc123"
}
```

## Authentication API

### Issue Token

Issues a new JWT access token.

**Endpoint:** `POST /auth/token`

**Request:**
```json
{
  "subject": "user-123",
  "roles": ["admin", "user"],
  "audience": "api",
  "metadata": {
    "service": "payment-service"
  }
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "type": "access",
  "expires_at": "2025-12-15T10:00:00Z",
  "issued_at": "2025-12-15T09:00:00Z"
}
```

### Validate Token

Validates a JWT token and returns its claims.

**Endpoint:** `POST /auth/validate`

**Request:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `200 OK`
```json
{
  "sub": "user-123",
  "iss": "root-server",
  "aud": "api",
  "exp": "2025-12-15T10:00:00Z",
  "iat": "2025-12-15T09:00:00Z",
  "roles": ["admin", "user"],
  "metadata": {
    "service": "payment-service"
  }
}
```

### Refresh Token

Generates a new access token from a refresh token.

**Endpoint:** `POST /auth/refresh`

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "type": "access",
  "expires_at": "2025-12-15T11:00:00Z",
  "issued_at": "2025-12-15T10:00:00Z"
}
```

### Revoke Token

Revokes a token, adding it to the blacklist.

**Endpoint:** `POST /auth/revoke`

**Request:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `204 No Content`

## Session Management API

### Create Session

Creates a new session.

**Endpoint:** `POST /session`

**Request:**
```json
{
  "user_id": "user-123",
  "service_id": "payment-service",
  "data": {
    "cart_id": "cart-456",
    "preferences": {
      "currency": "USD"
    }
  },
  "ttl": 60
}
```

**Response:** `201 Created`
```json
{
  "id": "sess_1234567890",
  "user_id": "user-123",
  "service_id": "payment-service",
  "data": {
    "cart_id": "cart-456",
    "preferences": {
      "currency": "USD"
    }
  },
  "created_at": "2025-12-15T09:00:00Z",
  "expires_at": "2025-12-15T10:00:00Z",
  "updated_at": "2025-12-15T09:00:00Z"
}
```

### Get Session

Retrieves a session by ID.

**Endpoint:** `GET /session/:id`

**Response:** `200 OK`
```json
{
  "id": "sess_1234567890",
  "user_id": "user-123",
  "service_id": "payment-service",
  "data": { ... },
  "created_at": "2025-12-15T09:00:00Z",
  "expires_at": "2025-12-15T10:00:00Z",
  "updated_at": "2025-12-15T09:00:00Z"
}
```

### Update Session

Updates session data.

**Endpoint:** `PUT /session/:id`

**Request:**
```json
{
  "data": {
    "cart_id": "cart-789",
    "preferences": {
      "currency": "EUR"
    }
  }
}
```

**Response:** `204 No Content`

### Delete Session

Deletes a session.

**Endpoint:** `DELETE /session/:id`

**Response:** `204 No Content`

## Service Registry API

### Register Service

Registers a new service with the root server.

**Endpoint:** `POST /registry/register`

**Request:**
```json
{
  "id": "payment-svc-1",
  "name": "payment-service",
  "version": "1.2.0",
  "endpoints": [
    "http://payment-1:8080",
    "http://payment-2:8080"
  ],
  "capabilities": ["payment", "refund", "subscription"],
  "metadata": {
    "region": "us-east-1",
    "environment": "production"
  },
  "health_check_url": "http://payment-1:8080/health"
}
```

**Response:** `201 Created`
```json
{
  "id": "payment-svc-1",
  "name": "payment-service",
  "version": "1.2.0",
  "endpoints": [...],
  "capabilities": [...],
  "metadata": {...},
  "status": "healthy",
  "registered_at": "2025-12-15T09:00:00Z",
  "last_heartbeat": "2025-12-15T09:00:00Z",
  "health_check_url": "http://payment-1:8080/health"
}
```

### Deregister Service

Removes a service from the registry.

**Endpoint:** `DELETE /registry/deregister/:id`

**Response:** `204 No Content`

### List Services

Returns all registered services.

**Endpoint:** `GET /registry/services`

**Response:** `200 OK`
```json
[
  {
    "id": "payment-svc-1",
    "name": "payment-service",
    "version": "1.2.0",
    "status": "healthy",
    ...
  },
  {
    "id": "notification-svc-1",
    "name": "notification-service",
    "version": "2.0.0",
    "status": "healthy",
    ...
  }
]
```

### Discover Services

Finds services by capability.

**Endpoint:** `GET /registry/discover?capability=payment`

**Query Parameters:**
- `capability` (optional): Filter by capability

**Response:** `200 OK`
```json
[
  {
    "id": "payment-svc-1",
    "name": "payment-service",
    "capabilities": ["payment", "refund"],
    "status": "healthy",
    ...
  }
]
```

### Send Heartbeat

Updates the heartbeat timestamp for a service.

**Endpoint:** `PUT /registry/heartbeat/:id`

**Response:** `204 No Content`

## Health Check API

### Liveness Probe

Checks if the server is alive.

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "ok"
}
```

### Readiness Probe

Checks if the server is ready to handle requests.

**Endpoint:** `GET /ready`

**Response:** `200 OK`
```json
{
  "status": "ready"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_REQUEST | 400 | Request body is malformed |
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource already exists |
| INTERNAL_ERROR | 500 | Internal server error |

## Rate Limiting

Rate limiting is applied per API key:

- 1000 requests per minute for normal operations
- 100 requests per minute for token issuance

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1639584000
```

## Versioning

API version is included in the Accept header:

```
Accept: application/vnd.root.v1+json
```

Current version: `v1` (default)
