FROM golang:1.16.2-alpine AS builder

WORKDIR /build

COPY consumer/go.* ./
RUN go mod download

COPY consumer/*.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer .


FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/consumer .

CMD ["./consumer"]