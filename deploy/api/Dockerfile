
FROM golang:1.26-alpine AS go-builder

RUN apk add --no-cache git make

WORKDIR /builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:latest

RUN apk add --no-cache docker-cli bash curl ca-certificates

COPY --from=go-builder /builder/api /api
RUN chmod +x /api

ENTRYPOINT ["/api"]