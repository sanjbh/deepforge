FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o deepforge ./cmd/deepforge

FROM alpine:latest
ENV DEEPFORGE_QUERY="default research topic"
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/deepforge ./
# CMD ./deepforge -query="${DEEPFORGE_QUERY}"
CMD ["/bin/sh", "-c", "./deepforge -query=\"${DEEPFORGE_QUERY}\""]
