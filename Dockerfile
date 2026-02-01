FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o main ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

RUN chown -R app:app /app

USER app

EXPOSE 8080

CMD ["./main"]