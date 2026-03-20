FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
ENV GOPROXY=direct
ENV GONOSUMDB=github.com/robert7528/hycore
ENV GONOSUMCHECK=github.com/robert7528/hycore
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o hyadmin ./cmd/hyadmin

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/hyadmin .
COPY configs/ configs/
COPY migrations/ migrations/
COPY deployment/entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh

RUN mkdir -p logs

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
