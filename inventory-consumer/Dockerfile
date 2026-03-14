FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY inventory-consumer/go.mod inventory-consumer/go.sum ./inventory-consumer/
WORKDIR /app/inventory-consumer
RUN go mod download
COPY inventory-consumer/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o inventory-consumer ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/inventory-consumer/inventory-consumer .
ENTRYPOINT ["./inventory-consumer"]
