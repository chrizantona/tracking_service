FROM golang:1.23.8-alpine3.21 AS builder
RUN apk --no-cache add git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["/app/main"]
