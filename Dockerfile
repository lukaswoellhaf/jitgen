FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /jitgen ./cmd/jitgen

FROM golang:1.25-alpine
RUN apk add --no-cache git \
    && git config --system --add safe.directory /github/workspace \
    && git config --system user.name "jitgen" \
    && git config --system user.email "jitgen@users.noreply.github.com"
COPY --from=builder /jitgen /usr/local/bin/jitgen
ENTRYPOINT ["jitgen"]
