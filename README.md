# Silsilah Keluarga - Backend API

REST API for a family tree application built with Go.

## Features

- JWT authentication with session management
- Family member (person) and relationship management
- Consanguinity detection (blood relation calculation)
- Role-based access control (member, editor, developer)
- Change request approval workflow
- Media upload with MinIO (S3-compatible)
- Email notifications via Resend

## Tech Stack

Go • Chi Router • PostgreSQL • Redis • MinIO • JWT

## Quick Start

```bash
# Clone and setup
git clone https://github.com/yourusername/silsilah-be.git
cd silsilah-be
cp .env.example .env

# Configure .env with your database, redis, and minio settings

# Run migrations
migrate -path migrations -database "$DATABASE_URL" up

# Start server
go run cmd/api/main.go
```

Server runs at `http://localhost:8080`

## Project Structure

```
cmd/api/          # Application entry point
internal/
  ├── config/     # Configuration
  ├── domain/     # Domain models
  ├── handler/    # HTTP handlers
  ├── middleware/ # Auth, RBAC, error handling
  ├── repository/ # Database layer
  └── service/    # Business logic
migrations/       # SQL migrations
```

## Configuration

See [.env.example](.env.example) for all available environment variables.

## License

[MIT](LICENSE)
