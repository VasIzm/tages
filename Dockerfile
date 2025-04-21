FROM golang:1.23.4-alpine AS build

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main ./cmd/app

FROM alpine:3.21.0

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build app/main .
COPY --from=build app/.env .

CMD ["./main"]