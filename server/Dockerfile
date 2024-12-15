FROM golang:1.20
WORKDIR /app
COPY . .
RUN go build -o server main.go
CMD ["./server"]
