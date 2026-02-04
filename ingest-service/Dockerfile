FROM golang:1.21-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ingest ./ingest-service/cmd/ingest/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/ingest .
CMD ["./ingest"]
