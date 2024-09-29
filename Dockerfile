FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN apk --no-cache add gcc musl-dev sqlite-dev
ENV CGO_ENABLED=1
RUN go build -o main .


FROM alpine:latest

RUN apk --no-cache add sqlite sqlite-libs ffmpeg yt-dlp
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates

EXPOSE 8091

CMD ["./app"]
