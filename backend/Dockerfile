FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o /app/seed ./cmd/seed/main.go
RUN go build -o /app/server ./cmd/server/main.go
RUN go build -o /app/bot ./cmd/bot/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /app/server
COPY --from=builder /app/seed /app/seed
COPY --from=builder /app/bot /app/bot
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/server"]
