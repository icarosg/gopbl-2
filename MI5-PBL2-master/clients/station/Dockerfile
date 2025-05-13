FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/station
RUN go build -o station-client .

# RUN chmod +x station-client

EXPOSE 8080
CMD ["./station-client"]