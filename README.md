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

## 1. Clone the Repository

```bash
git clone <repository-url>
cd xm-companies-manager
```

## 2. Configure the Environment

Copy the provided example configuration:

```bash
cp .env.example .env
```

The default values are suitable for running the project locally with Docker Compose.

For local development, the default placeholder values can be used. To generate stronger secrets, run:

```bash
openssl rand -hex 32
```

Use generated values to replace `JWT_SECRET` and/or `DB_PASSWORD` in `.env`.

## 3. Start the Application

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

## 4. Generate a JWT Token

All API endpoints require JWT authentication.

Generate a test token by running:

```bash
make token
```

The command prints a JWT token.

Use it in requests as:

```text
Authorization: Bearer <token>
```

## 5. Try the API

You can test the API using either Postman or curl.

### Postman

A Postman collection is included in the repository:

```text
XM_Companies_Manager.postman_collection.json
```

Import it into Postman and configure the collection variable with the token generated in step 4:

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

### curl

Set the generated token:

```bash
export JWT="<generated-token>"
```

#### Create a Company

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

#### Get a Company

```bash
curl http://localhost:8001/companies/<company-id> \
  -H "Authorization: Bearer $JWT"
```

#### Patch a Company

```bash
curl -X PATCH http://localhost:8001/companies/<company-id> \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "amount_of_employees": 150,
    "registered": false
  }'
```

#### Delete a Company

```bash
curl -X DELETE http://localhost:8001/companies/<company-id> \
  -H "Authorization: Bearer $JWT"
```

## 6. Inspect Kafka Events

Mutating operations publish events to the Kafka topic:

```text
company-events
```

The following event types are produced:

```text
company.created
company.updated
company.deleted
```

After performing one or more mutating operations, inspect the produced events with:

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

## 7. Run the Tests

Integration tests use Testcontainers and start an isolated PostgreSQL instance automatically.

Docker must therefore be running.

```bash
make test
```

## 8. Install Development Tools

Install the configured version of `golangci-lint`:

```bash
make tools
```

## 9. Lint and Format

After installing the development tools, run the linter:

```bash
make lint
```

Format the Go code with:

```bash
make fmt
```

## 10. Architecture

```text
HTTP Request
    ↓
Gin / JWT Middleware
    ↓
Handler
    ↓
Service
   ├── Repository → sqlc → PostgreSQL
   └── Event Producer → Kafka
```

## 11. Development Resources

During the implementation of this exercise I used:

- Documentation for the libraries and tools used
- Google/search for documentation and troubleshooting
- ChatGPT for development assistance