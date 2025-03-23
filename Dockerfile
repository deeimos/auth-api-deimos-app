FROM golang:1.24 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o auth-api ./cmd/auth-api/main.go
RUN go build -o migrator ./cmd/migrator/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/auth-api .
COPY --from=builder /app/migrator .

VOLUME /app/config
EXPOSE 1000

HEALTHCHECK \
  --interval=10s \
  --timeout=3s \
  --start-period=5s \
  --retries=3 \
  CMD nc -z localhost 1000 || exit 1

CMD ["./auth-api"]