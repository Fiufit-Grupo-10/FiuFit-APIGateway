FROM golang:1.18.1-bullseye

WORKDIR /appp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY scripts/ ./scripts/

ARG FIREBASE
RUN echo "${FIREBASE}" >> firebase.json

COPY firebase.json .

RUN go build -o main ./cmd/main.go

ARG USERS_URL
ENV USERS_URL ${USERS_URL}

ENV GIN_MODE="release"
CMD ["./main"]

