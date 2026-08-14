FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o worker ./cmd/worker

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/worker .

CMD ["./worker"]
