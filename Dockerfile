# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o barber-backend .

# Final stage
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -u 1000 appuser

COPY --from=builder /app/barber-backend .

USER appuser
EXPOSE 3000
CMD ["./barber-backend"]
