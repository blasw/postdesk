FROM golang:1.22.5-alpine3.19 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o gateway 

FROM alpine:latest

WORKDIR /root/

COPY --from=build /app/gateway .

CMD ["./gateway"]
