FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hyadmin-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o hyadmin-migrate ./cmd/migrate

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/hyadmin-api .
COPY --from=builder /app/hyadmin-migrate .
COPY configs/ configs/
COPY migrations/ migrations/
COPY deployment/entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh

RUN mkdir -p logs

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
