FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o hyadmin-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o hyadmin-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o hyadmin-seed ./cmd/seed

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/hyadmin-api .
COPY --from=builder /app/hyadmin-migrate .
COPY --from=builder /app/hyadmin-seed .
COPY configs/ configs/
COPY migrations/ migrations/
COPY deployment/entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh

RUN mkdir -p logs

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
