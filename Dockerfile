# building
FROM golang:latest as builder
MAINTAINER Sergey A. Kutylev <sergey@kutylev.com>
ADD . /app/
WORKDIR /app
RUN go get github.com/gorilla/mux
RUN go get github.com/jackc/pgx/pgxpool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /bookapp .

# gopkg for dependency managment gopkg.toml, gopkg.lock
# packaging /etc/apk/repositories
FROM alpine:latest
COPY --from=builder /bookapp ./
EXPOSE 8080
ENTRYPOINT ["./bookapp"]
# multistaging building