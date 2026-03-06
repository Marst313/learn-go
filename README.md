# Golang REST API Learning Project

Proyek ini dibuat untuk belajar Golang melalui pembuatan REST API.

## 📋 Prerequisites

- Go 1.21 atau lebih tinggi
- Docker & Docker Compose (untuk deployment)
- PostgreSQL (atau gunakan Docker)
- Air (untuk hot reload development)

## 🚀 How to Run

### Development Mode (dengan hot reload)
```bash
# Install dependencies
go mod tidy

# Run dengan Air
air
# Atau: go run github.com/cosmtrek/air@latest
```

### Production Mode
```bash
# Build binary
go build -o bin/main ./cmd/api/main.go

# Run binary
./bin/main
```

### Dengan Docker
```bash
# Build dan run dengan Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Dengan Makefile
```bash
# Lihat semua commands
make help

# Development
make dev                         # Run dengan hot reload
make test                        # Run tests
make build                       # Build binary

# Docker
make compose-up                  # Start semua services
make compose-down                # Stop semua services
make compose-logs                # View logs
```

## 📁 Project Structure

```
/project-root
├── cmd/
│   ├── api/
│   │   └── main.go              # Application entry point
│   └── migrate/
│       └── migrations/          # Database migrations
│
├── internal/
│   ├── database/                # Database connection & setup
│   ├── middleware/              # Custom Gin middleware (auth, logging, etc)
│   ├── models/                  # Database models/entities
│   ├── repositories/            # Database operations (data layer)
│   ├── services/                # Business logic
│   ├── handlers/                # HTTP request handlers (controllers)
│   └── utils/                   # Helper functions
│
├── routes/                      # Route definitions
├── tmp/                         # Air temporary files (generated)
│
├── .env                         # Environment variables (DO NOT COMMIT!)
├── .env.example                 # Environment variables template
│
├── .github/
│   └── workflows/
│       └── deploy.yml           # CI/CD configuration
│
├── docker-compose.yml           # Docker services configuration
├── Dockerfile                   # Docker image definition
├── Makefile                     # Helper commands
│
├── .air.toml                    # Air configuration
├── .gitignore                   # Git ignore rules
├── .dockerignore                # Docker ignore rules
│
├── go.mod                       # Go dependencies
├── go.sum                       # Go dependencies checksum
│
└── README.md                    # This file
```

## 🔧 Configuration

### Environment Variables

Copy `.env.example` ke `.env` dan sesuaikan nilai-nilainya:

```bash
cp .env.example .env
```

```env
DATABASE_URL=*****
JWT_SECRET=******
PORT=8080
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests dengan coverage
make test-coverage

# Run specific test
go test ./internal/handlers -v
```

## 🐳 Docker

### Build Image
```bash
docker build -t golang-api .
```

### Run dengan Docker Compose
```bash
# Start semua services (app + database + redis)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f app

# Stop
docker-compose down
```

## 📦 Deployment

### Automatic Deployment (CI/CD)

Setiap push ke branch `main` akan otomatis:
1. Run tests
2. Build Docker image
3. Deploy ke VPS

**Setup GitHub Secrets:**
- `VPS_HOST` - IP address VPS
- `VPS_USERNAME` - SSH username
- `VPS_SSH_KEY` - Private SSH key
- `VPS_PORT` - SSH port (default: 22)

Lihat `.github/workflows/deploy.yml` untuk detail.

### Manual Deployment

```bash
# Di VPS, pull image terbaru
docker-compose pull

# Restart services
docker-compose up -d
```

## 📚 API Documentation

### Health Check
```bash
GET /health
Response: {"status":"healthy","timestamp":"2024-01-15T10:30:00Z"}
```

### API Endpoints
```
GET    /api/v1/users       # Get all users
GET    /api/v1/users/:id   # Get user by ID
POST   /api/v1/users       # Create user
PUT    /api/v1/users/:id   # Update user
DELETE /api/v1/users/:id   # Delete user
```

## 🛠️ Development

### Adding New Feature

```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Develop dengan hot reload
make dev

# 3. Write tests
# Edit file *_test.go

# 4. Run tests
make test

# 5. Commit
git add .
git commit -m "feat: add new feature"

# 6. Push
git push origin feature/new-feature

# 7. Create Pull Request di GitHub
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir cmd/migrate/migrations -seq create_users_table

# Run migrations
go run cmd/migrate/main.go up

# Rollback
go run cmd/migrate/main.go down
```

## 🐛 Troubleshooting

### Port already in use
```bash
# Linux/Mac
lsof -i :8080
kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Database connection failed
```bash
docker-compose ps db

docker-compose logs db

docker-compose restart db
```

### Hot reload not working
```bash
go install github.com/cosmtrek/air@latest

```

## 📖 Learn More

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM](https://gorm.io/docs/)
- [Docker Documentation](https://docs.docker.com/)

## 📝 License

This project is for learning purposes.

## 👥 Contributors

- I Nyoman Dharma

---
