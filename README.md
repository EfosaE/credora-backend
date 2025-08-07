# Credora Backend

A robust Go-based backend service for handling financial transactions, user management, and virtual account operations using the Monnify payment gateway.

## Features

### Core Functionality
- User Authentication & Authorization
  - JWT-based authentication
  - Secure password hashing
  - User registration and login

### Banking & Financial Features
- Virtual Account Management
  - Creation of reserved accounts via Monnify integration
  - Account deletion functionality
  - Multi-bank support
  
- Transaction Processing
  - Webhook handling for payment notifications
  - Transaction history tracking
  - Balance management and crediting
  
### Event System
- Redis-based event bus
- Event-driven architecture for account creation
- Asynchronous event processing

### Email Service
- Templated email notifications
- Welcome emails
- Account information emails

## Technical Stack

### Programming Language & Framework
- Go (Golang)
- Chi Router for HTTP routing
- SQLC for type-safe SQL

### Database
- PostgreSQL
- Database migrations using `golang-migrate`
- Redis for event bus

### Development Tools
- Make for build automation
- Docker for containerization
- Environment-based configuration

### Testing
- Unit testing with Go's testing package
- Mock implementations for external services
- Stub data for testing scenarios

## Project Structure

```
├── cmd/                  # Application entrypoint
├── domain/              # Business logic and interfaces
├── infrastructure/      # External service implementations
├── internal/            # Internal packages
│   ├── config/         # Configuration management
│   ├── db/             # Database operations
│   ├── handler/        # HTTP handlers
│   ├── router/         # Route definitions
│   └── server/         # Server setup
├── service/            # Business service implementations
└── test/               # Test utilities and mocks
```

## Design Patterns

1. **Repository Pattern**
   - Separation of data access logic
   - Interface-based design for flexibility

2. **Dependency Injection**
   - Constructor-based dependency injection
   - Clear service dependencies

3. **Service Layer**
   - Business logic encapsulation
   - Service-to-service communication

4. **Event-Driven Architecture**
   - Asynchronous processing
   - Loose coupling via events

5. **Middleware Pattern**
   - Authentication middleware
   - Logging and recovery

## Getting Started

### Prerequisites
- Go 1.x
- PostgreSQL
- Redis
- Make

### Setup

1. Clone the repository
2. Set up environment variables (copy `.env.example` to `.env`)
3. Run database migrations:
   ```bash
   make migrate-up
   ```

4. Generate SQLC code:
   ```bash
   make sqlc-generate
   ```

5. Start the server:
   ```bash
   make run
   ```

### Development Commands

- `make dev-setup` - Set up development environment
- `make test-app` - Run all tests
- `make build` - Build the application
- `make migrate-create` - Create new migration files

## Testing

The project includes comprehensive test coverage:

- Unit tests for services
- Integration tests for API endpoints
- Mock implementations for external dependencies
- Stub data for consistent test scenarios

