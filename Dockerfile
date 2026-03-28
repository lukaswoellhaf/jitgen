FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /jitgen ./cmd/jitgen

FROM alpine:3.21
RUN apk add --no-cache git \
    && git config --global --add safe.directory /github/workspace
COPY --from=builder /jitgen /usr/local/bin/jitgen
ENTRYPOINT ["jitgen"]
