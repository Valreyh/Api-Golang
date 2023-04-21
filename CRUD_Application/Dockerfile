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

RUN mkdir -p /app/images \
    mkdir -p /app/html_pages

COPY ./CRUD_Application/images /app/images
COPY ./CRUD_Application/html_pages /app/html_pages

WORKDIR /app
COPY --from=builder /build/main .

CMD ["./main"]
