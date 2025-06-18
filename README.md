# Credora Backend

This is the backend service for the Credora application.

## Project Structure

```
myproject/
├── cmd/                    # Main applications
│   └── server/            # The server application
├── internal/              # Private application code
├── pkg/                   # Public libraries
├── api/                   # API documentation
├── web/                   # Web assets
├── deployments/           # Deployment configurations
└── test/                  # Additional tests
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- PostgreSQL
- Make

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   make deps
   ```
3. Create and configure your `.env` file based on `.env.example`
4. Build the application:
   ```bash
   make build
   ```

### Running the Application

```bash
make run
```

### Running Tests

```bash
make test
```

## Development

- Use `make fmt` to format code
- Use `make lint` to run linters
- Use `make doc` to generate API documentation

## License

This project is licensed under the MIT License - see the LICENSE file for details.
