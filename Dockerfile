FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /scan-service ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /scan-service /scan-service

EXPOSE 8080
ENTRYPOINT ["/scan-service"]
