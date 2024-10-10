#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -v .
RUN go build -o /go/bin/app -v .

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
ENTRYPOINT ["/app","serve","--http=0.0.0.0:8090"]
VOLUME [ "/pb_data" ]
EXPOSE 8090

