FROM golang:1.22-alpine as builder
WORKDIR /app
COPY main.go .
RUN go build -o runner main.go

FROM alpine:latest
WORKDIR /code
COPY --from=builder /app/runner /runner
ENTRYPOINT ["/runner"] 