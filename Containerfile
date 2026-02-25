FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hyadmin-api ./cmd/server

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/hyadmin-api .
COPY configs/ configs/

RUN mkdir -p logs

EXPOSE 8080
CMD ["./hyadmin-api", "serve"]
