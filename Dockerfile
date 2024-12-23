FROM golang:1.23-alpine AS builder

RUN apk --no-cache add ca-certificates gcc g++ libc-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY ./configs/.env .env

RUN export $(cat .env | xargs) && go build -o backend ./cmd/main

FROM alpine:3.16

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/backend ./backend

CMD ["./backend"]
