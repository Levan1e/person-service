FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o person-service ./cmd/person-service

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/person-service .

COPY .env .

EXPOSE 8081

CMD ["./person-service"]