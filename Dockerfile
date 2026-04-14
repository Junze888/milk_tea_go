# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/milktea-server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
ENV TZ=Asia/Shanghai
COPY --from=build /out/milktea-server /app/milktea-server
COPY migrations /app/migrations
ENV MIGRATIONS_DIR=/app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/milktea-server"]
