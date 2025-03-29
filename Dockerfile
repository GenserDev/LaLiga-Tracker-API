FROM golang:1.24.1-alpine

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY *.go ./

RUN go build -o main .

EXPOSE 8081

CMD ["./main"]