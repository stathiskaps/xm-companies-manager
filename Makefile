BIN=../../bin/
API=./api/src/

run:
	docker compose up --build

tools:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.12.2

lint:
	cd $(API) && $(BIN)/golangci-lint run

fmt:
	cd $(API) && $(BIN)/golangci-lint fmt