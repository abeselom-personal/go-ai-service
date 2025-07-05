## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.24+
- PostgreSQL
- (Optional) Redis

### Clone the Repository

```bash
git clone https://github.com/abeselom-personal/go-ai-service.git
cd ai-prompt-service
````

### Run with Docker

```bash
docker-compose up --build
```

Access the API at:

* REST: `http://localhost:8082/api/...`
* WebSocket: `ws://localhost:8082/ws/...`
* gRPC: `localhost:8082` (adjust if needed)

### Run Locally (Air for Live Reload)

Install [Air](https://github.com/cosmtrek/air):

```bash
go install github.com/air-verse/air@latest
```

Run:

```bash
make run
```

Or manually:

```bash
air -c .air.toml
```

### Running Tests

```bash
make test
```

### Additional Commands

* `make logs` – tail container logs
* `make restart` – restart the service
* `make lint` – run linter (requires golangci-lint)
* `make fmt` – auto format Go code
* `make sh` – open shell in container

### Environment Setup

Create a `.env` file inside `ai-service/`:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=mydb
DATABASE_URL=postgres://postgres:postgres@ai-database:5432/mydb?sslmode=disable
REDIS_URL=redis://redis:6379
OPENAI_API_KEY=your_key
```

## License

MIT

```
