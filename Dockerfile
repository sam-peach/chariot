FROM --platform=linux/arm64 golang:1.21.1 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM --platform=linux/arm64 ubuntu:22.04

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
