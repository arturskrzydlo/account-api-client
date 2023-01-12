FROM golang:1.19-alpine

# name of the container with fake api in docker-compose
ENV ACCOUNT_API_HOSTNAME account-api

RUN apk add --no-cache make gcc libc-dev

WORKDIR /app

COPY . .

CMD make all-tests