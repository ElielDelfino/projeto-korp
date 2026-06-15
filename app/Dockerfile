FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /http-server-projeto-korp

FROM alpine:3.20

WORKDIR /

COPY --from=builder /http-server-projeto-korp /http-server-projeto-korp

EXPOSE 8080

ENTRYPOINT ["/http-server-projeto-korp"]