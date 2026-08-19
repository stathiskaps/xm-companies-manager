BIN=../bin/
API=api/

run:
	docker compose up --build

test:
	cd $(API) && go test ./...

tools:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.12.2

lint:
	cd $(API) && $(BIN)/golangci-lint run

fmt:
	cd $(API) && $(BIN)/golangci-lint fmt