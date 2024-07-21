FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download
RUN go get github.com/pressly/goose/v3/cmd/goose
RUN go install github.com/pressly/goose/v3/cmd/goose

