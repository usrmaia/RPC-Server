FROM golang:1.18

WORKDIR /app

COPY main.go ./
COPY go.mod ./
COPY go.sum ./
COPY pb ./pb/
COPY proto ./proto/

RUN ls /app
RUN go build -o server
EXPOSE 9091

CMD ["/app/server"]