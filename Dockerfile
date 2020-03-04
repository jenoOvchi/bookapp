FROM golang:latest as builder
MAINTAINER Sergey A. Kutylev <sergey@kutylev.com>
ADD . /app/
WORKDIR /app
RUN go get github.com/gorilla/mux
RUN go get github.com/jackc/pgx/pgxpool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /bookapp .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bookapp ./
ENTRYPOINT ["./bookapp"]
EXPOSE 8080