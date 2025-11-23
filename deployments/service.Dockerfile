FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/main.go

FROM alpine:latest AS runner

WORKDIR /

COPY --from=builder /service /service

EXPOSE 8080

ENTRYPOINT ["/service", "-config", "/app/config.yaml"]
