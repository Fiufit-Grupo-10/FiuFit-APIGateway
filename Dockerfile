#dev dockerfile
FROM golang:1.18.1-bullseye

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air"]
