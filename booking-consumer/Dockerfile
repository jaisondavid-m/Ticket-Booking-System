FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY shared/ ./shared/
COPY booking-consumer/go.mod booking-consumer/go.sum ./booking-consumer/
WORKDIR /app/booking-consumer
RUN go mod download
COPY booking-consumer/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o booking-consumer ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/booking-consumer/booking-consumer .
ENTRYPOINT ["./booking-consumer"]
