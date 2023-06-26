FROM golang:1.20-alpine AS builder

RUN apk add --no-cache git

COPY . /app
WORKDIR /app

RUN go mod tidy
RUN go build -o /app/build/updates_controller /app/cmd/updates_controller/main.go

FROM scratch

COPY --from=builder /app/build/updates_controller /updates_controller
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/updates_controller"]
