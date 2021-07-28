FROM golang:1.16.2-alpine AS builder

WORKDIR /build

COPY tester/go.* ./
RUN go mod download

COPY tester/*.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tester .


FROM alpine:latest

RUN apk add tzdata

WORKDIR /app

COPY --from=builder /build/tester .

CMD ["./tester"]