FROM golang:1.19.5
ENV PROJECT_DIR=avas-open-indexer

RUN mkdir /$PROJECT_DIR
WORKDIR /$PROJECT_DIR
COPY . .
RUN go build -o ./indexer ./cmd/main.go
