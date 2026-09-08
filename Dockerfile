# Build stage
FROM golang:1.27-alpine AS builder
WORKDIR /app

# Install the application dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the API executable and the migrations executable
COPY . .
RUN go build -o /gochat ./cmd/api
RUN CGO_ENABLED=0 go build -o /migration-executable ./cmd/migrations

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the compiled Go application from the builder
COPY --from=builder /gochat /app/gochat
COPY --from=builder /migration-executable /app/migration-executable
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/gochat"]
