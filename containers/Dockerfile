FROM golang:1.14

WORKDIR /app

COPY ./ ./

RUN GOOS=linux GOARCH=amd64 make build

CMD ["./node-selector-webhook"]

