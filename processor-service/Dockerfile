FROM golang:1.21-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o processor ./cmd/processor/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/processor .
CMD ["./processor"]
