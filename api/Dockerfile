FROM golang:1.26.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /company-api ./cmd/api

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1


FROM alpine:3.22

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -s /sbin/nologin appuser

WORKDIR /app

COPY --from=builder /company-api /usr/local/bin/company-api
COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY sql/migrations /app/sql/migrations
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

USER appuser

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]