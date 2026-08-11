FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/main ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./main"]
