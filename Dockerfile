FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN apk --no-cache add make
#ENV CGO_ENABLED=1
RUN make generate && go build -o main .


FROM alpine:latest

RUN apk --no-cache add curl python3 ffmpeg
WORKDIR /app
RUN mkdir -p ./.cache ./audio
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod +x /usr/local/bin/yt-dlp

EXPOSE 8091

CMD ["./main"]
