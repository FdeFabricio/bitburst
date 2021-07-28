FROM golang:1.16.2-alpine AS builder

WORKDIR /build

COPY routine/go.* ./
RUN go mod download

COPY routine/*.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o routine .


FROM alpine:latest

RUN apk add tzdata

WORKDIR /app

COPY --from=builder /build/routine .

CMD ["./routine"]