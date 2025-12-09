FROM golang:1.23.0-alpine3.20 AS builder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct

RUN mkdir -p /build
WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go mod tidy
RUN go build -ldflags="-s -w" -o /build/monitor .


FROM alpine:latest as run
RUN apk --no-cache add ca-certificates && rm -rf /var/cache/apk/*

ENV APP_PORT=8080

RUN mkdir -p /app
WORKDIR /app

# Copy the application executable from the build image
COPY --from=builder /build/monitor /app/monitor

EXPOSE $APP_PORT

ENTRYPOINT ["/app/monitor"]
