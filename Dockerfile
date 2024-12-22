FROM golang:1.23-alpine

RUN apk --no-cache add ca-certificates gcc g++ libc-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY ./configs/.env .env
RUN export $(cat .env | xargs) && go build -o backend ./cmd/main

#EXPOSE ${HTTP_SERVER_PORT}

CMD ["./backend"]
