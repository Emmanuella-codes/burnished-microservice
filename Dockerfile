# Step 1: Build stage
FROM golang:1.24.3-alpine3.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main .

# Step 2: Runtime stage
FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /app/main /app/main

USER app

EXPOSE 8080
CMD ["/app/main"]
