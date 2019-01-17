FROM golang:alpine as builder
RUN apk add git build-base
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o getter ./cmd/getter
FROM alpine
RUN adduser -S -D -H -h /app getteruser
USER getteruser
COPY --from=builder /build/getter /app/
WORKDIR /app
ENTRYPOINT [ "./getter" ]
CMD []
