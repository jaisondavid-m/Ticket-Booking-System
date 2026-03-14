FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /app/api-gateway
RUN go mod download
COPY api-gateway/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-gateway ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api-gateway/api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]