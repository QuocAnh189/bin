# Deployment Guide

## Prerequisites

- Docker and Docker Compose (optional)
- PostgreSQL 14+ (production)
- Redis 6+ (production)
- TLS certificates (production)

## Configuration

### Environment Variables

Required environment variables for production:

```bash
# Server
CONFIG_PATH=config/production/config.json

# JWT
JWT_SECRET=<your-secret-key>  # Use a strong random secret

# Redis
REDIS_ADDR=redis.internal:6379
REDIS_PASSWORD=<redis-password>

# PostgreSQL
POSTGRES_HOST=postgres.internal
POSTGRES_USER=rootserver
POSTGRES_PASSWORD=<postgres-password>
POSTGRES_DB=rootserver
```

### TLS Certificates

Place your TLS certificates in:

- Certificate: `/etc/ssl/certs/server.crt`
- Private Key: `/etc/ssl/private/server.key`

Or specify custom paths in `config/production/config.json`.

## Deployment Options

### Option 1: Docker Compose

Create `docker-compose.yml`:

```yaml
version: "3.8"

services:
  root-server:
    image: root-server:latest
    ports:
      - "443:443"
    environment:
      - CONFIG_PATH=config/production/config.json
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_ADDR=redis:6379
      - POSTGRES_HOST=postgres
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - ./config:/root/config
      - ./certs:/etc/ssl/certs
    depends_on:
      - redis
      - postgres
    restart: always

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
    restart: always

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=rootserver
      - POSTGRES_USER=rootserver
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    restart: always

volumes:
  redis-data:
  postgres-data:
```

Deploy:

```bash
docker-compose up -d
```

### Option 2: Kubernetes

Create deployment manifests:

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: root-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: root-server
  template:
    metadata:
      labels:
        app: root-server
    spec:
      containers:
        - name: root-server
          image: root-server:latest
          ports:
            - containerPort: 443
          env:
            - name: CONFIG_PATH
              value: config/production/config.json
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: root-server-secrets
                  key: jwt-secret
            - name: REDIS_ADDR
              value: redis:6379
            - name: POSTGRES_HOST
              value: postgres
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
          livenessProbe:
            httpGet:
              path: /health
              port: 443
              scheme: HTTPS
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: 443
              scheme: HTTPS
            initialDelaySeconds: 5
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: root-server
spec:
  selector:
    app: root-server
  ports:
    - port: 443
      targetPort: 443
  type: LoadBalancer
```

Deploy:

```bash
kubectl apply -f deployment.yaml
```

### Option 3: Systemd Service

Create `/etc/systemd/system/root-server.service`:

```ini
[Unit]
Description=Root Server
After=network.target

[Service]
Type=simple
User=rootserver
WorkingDirectory=/opt/root-server
Environment="CONFIG_PATH=/opt/root-server/config/production/config.json"
EnvironmentFile=/opt/root-server/.env
ExecStart=/opt/root-server/github.com/aq189/bin/rootserver
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable root-server
sudo systemctl start root-server
sudo systemctl status root-server
```

## Database Setup

### Run Migrations

```bash
# Install migrate tool
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations \
  -database "postgres://rootserver:password@localhost:5432/rootserver?sslmode=disable" \
  up
```

### Initialize Data

```sql
-- Create initial admin token (example)
INSERT INTO token_blacklist (token_id, revoked_at, expires_at)
VALUES ('initial', NOW(), NOW() + INTERVAL '1 year');
```

## Monitoring

### Health Checks

- Liveness: `GET /health`
- Readiness: `GET /ready`

### Metrics

Expose metrics for Prometheus:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: "root-server"
    static_configs:
      - targets: ["root-server:8080"]
```

### Logging

Logs are written to stdout in JSON format. Configure log aggregation:

```yaml
# filebeat.yml
filebeat.inputs:
  - type: container
    paths:
      - "/var/lib/docker/containers/*/*.log"
    processors:
      - add_docker_metadata: ~
```

## Backup

### PostgreSQL Backup

```bash
# Daily backup
pg_dump -h postgres -U rootserver rootserver | gzip > backup-$(date +%Y%m%d).sql.gz

# Restore
gunzip < backup-20251215.sql.gz | psql -h postgres -U rootserver rootserver
```

### Redis Backup

```bash
# Save snapshot
redis-cli BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb /backup/redis-$(date +%Y%m%d).rdb
```

## High Availability

### Load Balancer Configuration

```nginx
upstream root_servers {
    least_conn;
    server root-1.internal:443;
    server root-2.internal:443;
    server root-3.internal:443;
}

server {
    listen 443 ssl;
    server_name root.company.internal;

    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;

    location / {
        proxy_pass https://root_servers;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Request-ID $request_id;
    }
}
```

### Redis Sentinel

For Redis HA, configure Sentinel:

```
sentinel monitor mymaster redis-master 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 10000
```

### PostgreSQL Replication

Set up streaming replication for PostgreSQL:

```
# postgresql.conf (primary)
wal_level = replica
max_wal_senders = 3
```

## Troubleshooting

### Check Logs

```bash
# Docker
docker logs root-server

# Systemd
journalctl -u root-server -f

# Kubernetes
kubectl logs -f deployment/root-server
```

### Debug Mode

Enable debug logging:

```json
{
  "log": {
    "level": "debug",
    "format": "json"
  }
}
```

### Common Issues

**Issue:** Connection refused

- Check if server is running: `netstat -tlnp | grep 443`
- Verify firewall rules
- Check TLS certificates

**Issue:** Token validation fails

- Verify JWT_SECRET matches across instances
- Check token expiration
- Ensure clock synchronization (NTP)

**Issue:** High memory usage

- Check session cleanup is running
- Review Redis memory policy
- Monitor for connection leaks

## Security Checklist

- [ ] Strong JWT secret (>32 characters)
- [ ] TLS enabled with valid certificates
- [ ] Database credentials rotated
- [ ] Firewall rules configured
- [ ] Rate limiting enabled
- [ ] Audit logging enabled
- [ ] Regular security updates
- [ ] Backup encryption enabled
