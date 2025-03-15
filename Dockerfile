FROM golang:1.24.1-alpine

RUN apk update

RUN apk add --no-cache ffmpeg --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community 

RUN apk add --no-cache yt-dlp 

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -o out

ENTRYPOINT ["/app/out"]