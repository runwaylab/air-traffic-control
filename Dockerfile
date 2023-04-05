FROM golang:1.19.5-alpine3.16 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o main

FROM alpine:3.17.3 AS production

COPY --from=builder /app .

EXPOSE 8080
CMD ["./main"]
