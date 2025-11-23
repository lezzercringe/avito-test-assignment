EXECUTABLE_PATH := ./build/api
ENTRYPOINT_PATH := ./cmd/main.go
GOENV := GOEXPERIMENT=jsonv2

default: build

.PHONY: default codegen build

codegen:
	sqlc generate

build: codegen
	${GOENV} go build -o ${EXECUTABLE_PATH} ${ENTRYPOINT_PATH}

run:
	${GOENV} go run ${ENTRYPOINT_PATH}

lint:
	${GOENV} go tool staticcheck ./...
