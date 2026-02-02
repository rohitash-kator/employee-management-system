# Employee Management REST API (Gin + MongoDB)

Clean/Hexagonal-style architecture with:
- Domain entities + repository ports (`internal/domain`)
- Use cases / services (`internal/usecase`)
- Adapters (MongoDB repository) (`internal/adapters`)
- Delivery (Gin HTTP API) (`internal/delivery`)

## Prereqs

- Go 1.22+
- Docker (for local MongoDB)

## Run locally

1. Start MongoDB:

```bash
make mongo-up
```

2. Create your env file:

```bash
cp .env.example .env
```

3. Run the API:

```bash
make run
```

Health check:

```bash
curl http://localhost:8080/healthz
```

## REST Endpoints

Base path: `/v1`

- `POST /v1/employees` - create employee
- `GET /v1/employees` - list employees (supports `limit`, `offset`, `department`, `status`, `q`)
- `GET /v1/employees/:id` - get employee by id
- `PATCH /v1/employees/:id` - partial update
- `DELETE /v1/employees/:id` - delete

### Example: create

```bash
curl -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Rohit",
    "last_name": "Sharma",
    "email": "rohit@example.com",
    "department": "Engineering",
    "position": "Backend Engineer",
    "salary": 120000,
    "status": "active"
  }'
```

### Example: list

```bash
curl "http://localhost:8080/v1/employees?limit=20&offset=0&department=Engineering&status=active&q=rohit"
```

