FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /sentinel-api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

COPY --from=builder /sentinel-api .

EXPOSE 8090

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget --spider -q http://localhost:8090/health || exit 1

CMD ["./sentinel-api"]
