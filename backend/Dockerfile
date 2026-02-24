FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o steamzip-api .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/steamzip-api /usr/local/bin/steamzip-api

EXPOSE 8080

CMD ["steamzip-api"]
