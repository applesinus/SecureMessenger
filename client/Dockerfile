FROM golang:1.20
WORKDIR /app
COPY . .
RUN go build -o client main.go
CMD ["./client"]
