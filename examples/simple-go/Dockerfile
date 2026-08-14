FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .

# Resolve dependencies and generate go.sum for the templated go.mod
RUN go mod tidy

# Build the application
RUN go build -o worker ./cmd/worker

FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/worker .

CMD ["./worker"]
