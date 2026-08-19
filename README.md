# XM Companies Manager

A Go microservice for managing companies, implemented as part of the XM Golang technical exercise.

The service supports:

- Create a company
- Get a company
- Patch a company
- Delete a company
- JWT authentication
- PostgreSQL persistence
- Kafka events for mutating operations
- Database migrations with Goose
- SQL generation with sqlc
- Integration tests using Testcontainers
- Docker Compose setup
- golangci-lint

## Requirements

To run the application:

- Docker
- Docker Compose

To run tests, generate JWT tokens, or execute the linter locally:

- Go
- Make

## 1. Clone the repository

```bash
git clone <repository-url>
cd xm-companies-manager
```

## 2. Configure the environment

Copy the provided example configuration:

```bash
cp .env.example .env
```

The default values are suitable for running the project locally with Docker Compose.

For local development, the default secret values can be used. To generate stronger secrets, run:

```bash
openssl rand -hex 32
```

and replace:

```env
JWT_SECRET=change-me
```
and

```env
DB_PASSWORD=change-me
```

in `.env`.

## 3. Start the application

```bash
make run
```

This starts:

- Companies API
- PostgreSQL
- Kafka

Database migrations are applied automatically when the API starts.

The API is available by default at:

```text
http://localhost:8001
```

## 4. Generate a JWT token

All API endpoints require JWT authentication. Generate a test token by running:

```bash
make token
```

The command prints a JWT token.

Use it in requests as:

```text
Authorization: Bearer <token>
```

## 5. Try the API

You can testg the api by either using Postman or curl.

### 1. Postman

A Postman collection is included in the repository:

```text
XM_Companies_Manager.postman_collection.json
```

Import it into Postman and configure the collection variable with the token you received at step 4:

```text
jwt = <generated-token>
```

The collection includes requests for:

```text
POST   /companies/
GET    /companies/:companyId
PATCH  /companies/:companyId
DELETE /companies/:companyId
```

The Create Company request automatically stores the returned company ID for use by the other requests.

### 2. curl

Set the generated token you got at step 4:

```bash
export JWT="<generated-token>"
```

#### Create a company

```bash
curl -X POST http://localhost:8001/companies/ \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "XM Company",
    "description": "Example company",
    "amount_of_employees": 120,
    "registered": true,
    "type": "Corporations"
  }'
```

The response contains the generated company UUID.

#### Get a company

```bash
curl http://localhost:8001/companies/<company-id> \
  -H "Authorization: Bearer $JWT"
```

#### Patch a company

```bash
curl -X PATCH http://localhost:8001/companies/<company-id> \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_of_employees": 150,
    "registered": false
  }'
```

#### Delete a company

```bash
curl -X DELETE http://localhost:8001/companies/<company-id> \
  -H "Authorization: Bearer $JWT"
```

## 6. Inspect Kafka events

Mutating operations publish events to Kafka topic:

```text
company-events
```

The following event types are produced:

```text
company.created
company.updated
company.deleted
```

After performing one or more mutating operations, inspect the produced Kafka events with:

```bash
docker compose exec kafka \
  /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server kafka:9092 \
  --topic company-events \
  --from-beginning
```

Example output:

```json
{
  "id": "...",
  "type": "company.created",
  "company_id": "...",
  "timestamp": "..."
}
```

## 7. Run the tests

Integration tests use Testcontainers and start an isolated PostgreSQL instance automatically.

Docker must therefore be running.

```bash
make test
```

## 8. Install development tools

Install the configured version of `golangci-lint`:

```bash
make tools
```

## 9. Run the linter (step 8 is required)

```bash
make lint
```

Format the Go code with:

```bash
make fmt
```

## 10. Abstract Architecture

HTTP Request
    ↓
Gin / JWT Middleware
    ↓
Handler
    ↓
Service
   ├── Repository → sqlc → PostgreSQL
   └── Event Producer → Kafka

## 11. Development Resources

During the implementation of this exercise I used:

- Documentation and tools/libraries familiar from previous projects
- Google/search for documentation and troubleshooting
- ChatGPT for development assistance