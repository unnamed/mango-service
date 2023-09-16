# syntax=docker/dockerfile:1

FROM golang:1.21.1 AS go

WORKDIR /mango

COPY . .
RUN CG0_ENABLED=0 GOOS=linux go build -o mango-service

EXPOSE 2069

CMD ["./mango-service"]