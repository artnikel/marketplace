# Marketplace

A Go-based marketplace application with PostgreSQL database and web interface.

## Live Demo

The application is hosted at: https://marketplace-3aiq.onrender.com

## Technologies Used

- **Backend**: Go 1.24.2 with Gorilla Mux framework
- **Database**: PostgreSQL 15
- **Database Migrations**: Flyway
- **Authentication**: JWT tokens
- **Containerization**: Docker & Docker Compose
- **Testing**: Unit tests, mock testing, and integration tests with dockertest

## Features

- User registration and authentication
- JWT-based authorization
- Item management (create, list)
- RESTful API endpoints
- CORS support
- Health check endpoint
- Static file serving for frontend

---

## API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/items` - Get all items

### Protected Endpoints
- `POST /api/items` - Create new item (requires authentication)

---

## Local Development

### Prerequisites

- Docker
- Docker Compose

### Running with Docker Compose

1. Clone the repository
2. Run the application:

```bash
docker-compose up --build
```

This will start:
- PostgreSQL database on port 5432
- Flyway migrations
- Application server on port 8080

The application will be available at `http://localhost:8080`

### Configuration

The application uses a `config.yaml` file for configuration. 

### Database

The application uses PostgreSQL with Flyway for database migrations. Migration files should be placed in the `./migrations` directory.

## Testing

The project includes comprehensive testing:

- **Unit Tests**
- **Integration Tests**
- **Mock Testing**

Check tests cover with:
```bash
go test ./... -cover
```

## Code Quality

The project uses linting tools to maintain code quality. Make sure to run the linter before submitting changes:
```bash
golangci-lint run ./... --config=./.golangci.yml
```

## Project Structure

- `/internal` , `/pkg`- Application source code
- `/web` - Frontend static files
- `/migrations` - Database migration files
- `docker-compose.yml` - Docker Compose configuration
- `Dockerfile` - Multi-stage Docker build configuration
- `config.yaml` - Application configuration

