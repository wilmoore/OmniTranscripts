FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o omnitranscripts

FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ffmpeg \
    python3 \
    py3-pip \
    && pip3 install yt-dlp

# Install whisper.cpp
RUN apk add --no-cache git make g++ \
    && git clone https://github.com/ggerganov/whisper.cpp.git /tmp/whisper \
    && cd /tmp/whisper \
    && make \
    && cp main /usr/local/bin/whisper.cpp \
    && bash ./models/download-ggml-model.sh base.en \
    && mkdir -p /usr/local/share/whisper \
    && cp models/ggml-base.en.bin /usr/local/share/whisper/ \
    && rm -rf /tmp/whisper

WORKDIR /app
COPY --from=builder /app/omnitranscripts .
COPY .env .

EXPOSE 3000

CMD ["./omnitranscripts"]