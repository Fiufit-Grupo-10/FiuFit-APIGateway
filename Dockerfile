#dev dockerfile
FROM golang:1.18.1-bullseye

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

ARG USERS_URL
ENV USERS_URL ${USERS_URL}

