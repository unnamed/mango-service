# syntax=docker/dockerfile:1

FROM golang:1.21.1 AS go
WORKDIR /mango
COPY . .
RUN go build -o mango-service

FROM debian:buster
WORKDIR /mango
COPY --from=go /mango/mango-service .

CMD ["./mango-service"]