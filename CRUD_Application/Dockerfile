FROM golang:1.18 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY ./CRUD_Application .
RUN go build -o main ./cmd/*.go

EXPOSE 8080

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/main .

CMD ["./main"]
