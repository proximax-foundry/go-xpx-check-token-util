FROM golang:1.21.5-alpine3.19 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN apk add --no-cache build-base
COPY . ./
ENV TELEGRAM_APITOKEN=telegram_key
RUN CGO_ENABLED=1 GOOS=linux go build -o /check-token-util

FROM alpine:3.19
WORKDIR /
COPY --from=build /check-token-util .
COPY --from=build /app/config.json .
ENTRYPOINT ["/check-token-util"]