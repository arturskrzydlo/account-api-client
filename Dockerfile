FROM golang:1.19-alpine

RUN apk add --no-cache make gcc libc-dev

WORKDIR /app

COPY . .

CMD make all-tests