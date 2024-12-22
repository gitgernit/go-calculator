# go-calculator
Project for the Yandex.Lyceum Go programming course.  
The project implements the Shunting Yard algorithm to evaluate infix mathematical expressions.
An HTTP entrypoint to the calculator is provided alongside the logic.

## Prerequisites
* go
* Docker (optionally)

## Running
2 options are present, stick to the one that fits you more:

1. Run directly through go: `go run cmd/main/main.go`
2. Use docker with compose plugin: `docker compose --env-file=./configs/.env up --build`

Docker compose will expect you to have some environment variables, 
hence, you'll need to create an .env file or export them manually. 
Feel free to create the .env file based on ./configs/.env.template

## Usage
Use any tool you want to send requests to the server. For example,
you can use curl to send a POST request to api/v1/calculate:
```shell
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
"expression": "2+2*2"
}'
```
The server will respond with HTTP 422 to any kind of malformed \ invalid request. 
Allowed input data for expression is rational numbers (tested on integers and floats), 
prioritisation operators and basic math operators (+, -, *, /).
Deviations from allowed input data are considered invalid request, i.e.:
```shell
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
"expression": "2+2*2+"
}'
```
