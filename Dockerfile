# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/toy-blockchain \
    .

# Runtime stage
FROM alpine:3.20

RUN adduser -D -u 10001 blockchain

WORKDIR /app

COPY --from=builder /out/toy-blockchain ./toy-blockchain

USER blockchain

EXPOSE 8001 8002 8003

ENTRYPOINT ["./toy-blockchain"]