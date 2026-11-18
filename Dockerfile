# Multi-stage build: compile the CGO-based (go-sqlite3) binary on Alpine with
# musl, then run it in a minimal Alpine image. Per the PRD: one Go binary,
# one SQLite file, one Fly.io app/machine.
FROM golang:1.24-alpine AS build
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/menuapp .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/menuapp /app/menuapp

ENV MENUAPP_ADDR=:8080
ENV MENUAPP_DB=/data/menuapp.db

EXPOSE 8080
VOLUME ["/data"]

CMD ["/app/menuapp"]
