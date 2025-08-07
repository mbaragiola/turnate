# Turnate Deployment Guide

This guide covers various deployment options for Turnate, from development to production.

## 🔧 Development Deployment

### Local Development
```bash
# Clone and setup
git clone <repository-url>
cd turnate
go mod tidy

# Run in development mode
export JWT_SECRET=dev-secret-key
export PORT=8080
go run ./cmd/turnate
```

### Hot Reload Development
Using [Air](https://github.com/air-verse/air) for automatic reloading:

```bash
# Install Air
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

Create `.air.toml` for custom configuration:
```toml
root = "."
cmd = "go build -o ./tmp/main ./cmd/turnate"
bin = "tmp/main"

[build]
  exclude_dir = ["tmp", "vendor", "web/static", "tests"]
  include_ext = ["go", "html"]
  exclude_unchanged = true
  follow_symlink = true
```

## 🏭 Production Deployment

### 1. Binary Deployment

#### Build for Production
```bash
# Build optimized binary
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
  -a -ldflags '-extldflags "-static"' \
  -o bin/turnate ./cmd/turnate

# Create directory structure
mkdir -p /opt/turnate/{bin,data,logs,web}
cp bin/turnate /opt/turnate/bin/
cp -r web/* /opt/turnate/web/
```

#### Environment Configuration
Create `/opt/turnate/.env`:
```bash
PORT=8080
DATABASE_URL=/opt/turnate/data/turnate.db
JWT_SECRET=$(openssl rand -base64 32)
GIN_MODE=release
```

#### SystemD Service
Create `/etc/systemd/system/turnate.service`:
```ini
[Unit]
Description=Turnate Chat Server
After=network.target

[Service]
Type=simple
User=turnate
Group=turnate
WorkingDirectory=/opt/turnate
ExecStart=/opt/turnate/bin/turnate
EnvironmentFile=/opt/turnate/.env
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/turnate/data /opt/turnate/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

#### Setup and Start
```bash
# Create user
sudo useradd -r -s /bin/false turnate
sudo chown -R turnate:turnate /opt/turnate

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable turnate
sudo systemctl start turnate

# Check status
sudo systemctl status turnate
sudo journalctl -u turnate -f
```

### 2. Docker Deployment

#### Dockerfile
Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o turnate ./cmd/turnate

# Runtime stage  
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Create directories
RUN mkdir -p /app/{data,web}

# Copy binary and web assets
COPY --from=builder /app/turnate /app/
COPY --from=builder /app/web /app/web/

# Create non-root user
RUN adduser -D -s /bin/sh turnate
RUN chown -R turnate:turnate /app
USER turnate

WORKDIR /app
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./turnate"]
```

#### Docker Compose
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  turnate:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=/app/data/turnate.db
      - JWT_SECRET=your-super-secure-jwt-secret-here
      - GIN_MODE=release
    volumes:
      - turnate_data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  turnate_data:
```

#### Deploy with Docker
```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f turnate

# Scale (if needed)
docker-compose up -d --scale turnate=3
```

### 3. Reverse Proxy Setup

#### Nginx Configuration
Create `/etc/nginx/sites-available/turnate`:
```nginx
upstream turnate {
    server 127.0.0.1:8080;
    # Add more servers for load balancing
    # server 127.0.0.1:8081;
    # server 127.0.0.1:8082;
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=turnate_api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=turnate_auth:10m rate=5r/m;

server {
    listen 80;
    server_name your-domain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Static assets caching
    location /static/ {
        proxy_pass http://turnate;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # API rate limiting
    location /api/v1/auth/ {
        limit_req zone=turnate_auth burst=10 nodelay;
        proxy_pass http://turnate;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/ {
        limit_req zone=turnate_api burst=20 nodelay;
        proxy_pass http://turnate;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Main application
    location / {
        proxy_pass http://turnate;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support (if added later)
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Server $host;
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }

    # Health check
    location /health {
        proxy_pass http://turnate;
        access_log off;
    }
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/turnate /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### SSL with Let's Encrypt
```bash
# Install Certbot
sudo apt update
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal (crontab)
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

## ☁️ Cloud Deployment

### AWS EC2 Deployment

#### Launch Instance
1. Launch Ubuntu 22.04 LTS instance (t3.micro for testing)
2. Configure Security Group:
   - HTTP (80) - Source: 0.0.0.0/0
   - HTTPS (443) - Source: 0.0.0.0/0
   - SSH (22) - Source: Your IP

#### Setup Script
```bash
#!/bin/bash
# Install dependencies
sudo apt update
sudo apt install -y nginx certbot python3-certbot-nginx

# Download and setup Turnate
wget -O turnate.tar.gz <release-url>
sudo mkdir -p /opt/turnate
sudo tar -xzf turnate.tar.gz -C /opt/turnate
sudo useradd -r -s /bin/false turnate
sudo chown -R turnate:turnate /opt/turnate

# Configure environment
sudo tee /opt/turnate/.env > /dev/null <<EOF
PORT=8080
DATABASE_URL=/opt/turnate/data/turnate.db
JWT_SECRET=$(openssl rand -base64 32)
GIN_MODE=release
EOF

# Setup systemd service (see above)
# Setup nginx (see above)

# Start services
sudo systemctl enable turnate nginx
sudo systemctl start turnate nginx
```

### Digital Ocean Droplet

#### One-Click Setup Script
```bash
#!/bin/bash
set -e

echo "🚀 Installing Turnate on Digital Ocean..."

# Update system
apt update && apt upgrade -y

# Install dependencies
apt install -y wget nginx certbot python3-certbot-nginx

# Download Turnate
TURNATE_VERSION="latest"
wget -O /tmp/turnate.tar.gz "https://github.com/yourorg/turnate/releases/download/${TURNATE_VERSION}/turnate-linux-amd64.tar.gz"

# Setup directories
mkdir -p /opt/turnate/{bin,data,web,logs}
tar -xzf /tmp/turnate.tar.gz -C /opt/turnate/bin/
cp -r /tmp/web/* /opt/turnate/web/

# Create user
useradd -r -s /bin/false -d /opt/turnate turnate
chown -R turnate:turnate /opt/turnate

# Generate secure JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Create environment file
cat > /opt/turnate/.env <<EOF
PORT=8080
DATABASE_URL=/opt/turnate/data/turnate.db
JWT_SECRET=${JWT_SECRET}
GIN_MODE=release
EOF

# Create systemd service (contents from above)
# Create nginx config (contents from above)

# Enable services
systemctl daemon-reload
systemctl enable turnate nginx
systemctl start turnate nginx

# Setup firewall
ufw allow OpenSSH
ufw allow 'Nginx Full'
ufw --force enable

echo "✅ Turnate installed successfully!"
echo "🌐 Access your chat at: http://$(curl -s ifconfig.me)"
echo "👤 Default admin: admin / admin123"
echo "⚠️  Change the admin password immediately!"
```

### Kubernetes Deployment

#### Deployment YAML
Create `k8s/turnate-deployment.yaml`:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: turnate
  labels:
    app: turnate
spec:
  replicas: 3
  selector:
    matchLabels:
      app: turnate
  template:
    metadata:
      labels:
        app: turnate
    spec:
      containers:
      - name: turnate
        image: turnate:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: DATABASE_URL
          value: "/app/data/turnate.db"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: turnate-secrets
              key: jwt-secret
        - name: GIN_MODE
          value: "release"
        volumeMounts:
        - name: data-storage
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
      volumes:
      - name: data-storage
        persistentVolumeClaim:
          claimName: turnate-data
---
apiVersion: v1
kind: Service
metadata:
  name: turnate-service
spec:
  selector:
    app: turnate
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: turnate-data
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

#### Secrets and Ingress
```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: turnate-secrets
type: Opaque
data:
  jwt-secret: <base64-encoded-secret>
---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: turnate-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: turnate-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: turnate-service
            port:
              number: 80
```

Deploy:
```bash
kubectl apply -f k8s/
kubectl get pods -l app=turnate
kubectl logs -l app=turnate -f
```

## 📊 Monitoring & Logging

### System Monitoring
```bash
# Monitor system resources
top -p $(pgrep turnate)
iotop -p $(pgrep turnate)

# Check logs
journalctl -u turnate -f
tail -f /opt/turnate/logs/app.log
```

### Health Checks
Turnate exposes a health endpoint:
```bash
curl http://localhost:8080/health
# Expected response: {"status":"healthy","service":"turnate"}
```

### Log Rotation
Create `/etc/logrotate.d/turnate`:
```
/opt/turnate/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 turnate turnate
    postrotate
        systemctl reload turnate
    endscript
}
```

## 🔒 Security Considerations

### Production Security Checklist
- [ ] Use strong, randomly generated JWT secret
- [ ] Enable HTTPS with valid SSL certificates
- [ ] Configure firewall (UFW/iptables)
- [ ] Run as non-root user
- [ ] Enable fail2ban for SSH protection
- [ ] Regular security updates
- [ ] Database backups
- [ ] Rate limiting configured
- [ ] Security headers enabled

### Backup Strategy
```bash
#!/bin/bash
# backup-turnate.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups/turnate"
DB_PATH="/opt/turnate/data/turnate.db"

mkdir -p $BACKUP_DIR

# Backup database
sqlite3 $DB_PATH ".backup ${BACKUP_DIR}/turnate_${DATE}.db"

# Compress and keep only last 30 days
gzip ${BACKUP_DIR}/turnate_${DATE}.db
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete

echo "Backup completed: turnate_${DATE}.db.gz"
```

Add to crontab:
```bash
0 2 * * * /opt/scripts/backup-turnate.sh
```

## 📈 Performance Tuning

### Go Runtime Tuning
```bash
# Set environment variables
export GOMAXPROCS=4
export GOGC=100
export GOMEMLIMIT=256MiB
```

### Database Optimization
SQLite pragmas in application:
```go
// Add to database connection
db.Exec("PRAGMA journal_mode=WAL;")
db.Exec("PRAGMA synchronous=NORMAL;") 
db.Exec("PRAGMA cache_size=10000;")
db.Exec("PRAGMA temp_store=memory;")
```

### Nginx Optimization
```nginx
# Add to nginx config
worker_processes auto;
worker_connections 2048;

# Buffer sizes
client_body_buffer_size 128k;
client_max_body_size 10m;
client_header_buffer_size 1k;
large_client_header_buffers 4 4k;
output_buffers 1 32k;
postpone_output 1460;

# Timeouts  
client_body_timeout 12;
client_header_timeout 12;
keepalive_timeout 15;
send_timeout 10;

# Caching
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=turnate_cache:10m max_size=1g inactive=60m use_temp_path=off;
```

## 🚨 Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Find process using port
sudo lsof -i :8080
sudo kill -9 <PID>
```

#### Database Locked
```bash
# Check for stale locks
ls -la /opt/turnate/data/turnate.db-wal
# Remove if service is stopped
rm /opt/turnate/data/turnate.db-wal
```

#### High Memory Usage
```bash
# Monitor memory
ps aux | grep turnate
# Adjust GOMEMLIMIT if needed
```

#### SSL Certificate Issues
```bash
# Test certificate
openssl x509 -in /etc/letsencrypt/live/your-domain.com/fullchain.pem -text -noout
# Renew if needed
sudo certbot renew --force-renewal
```

### Debugging
```bash
# Enable debug mode
export GIN_MODE=debug
export LOG_LEVEL=debug

# Run with verbose logging
./turnate -v

# Check application logs
tail -f /opt/turnate/logs/app.log
```

This deployment guide should cover most scenarios from development to production-ready deployments. Choose the method that best fits your infrastructure and requirements.