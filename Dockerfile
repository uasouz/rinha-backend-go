FROM golang:1.21-bookworm as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /bin/api


FROM alpine:3.14

COPY --from=builder /bin/api /bin/api

CMD ["/bin/api"]