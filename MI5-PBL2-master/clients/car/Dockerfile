FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/car
RUN go build -o car-client .

EXPOSE 8080
CMD ["./car-client"]