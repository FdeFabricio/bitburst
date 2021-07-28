FROM golang:1.16.2-alpine AS builder

WORKDIR /build

COPY producer/go.* ./
RUN go mod download

COPY producer/*.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o producer .


FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/producer .

CMD ["./producer"]