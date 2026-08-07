# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=build /out/server /app/server
COPY data /app/data
COPY migrations /app/migrations
RUN mkdir -p /app/uploads/avatars && chown -R appuser:appuser /app
USER appuser
ENV ADDR=:8080 \
    COUNTRIES_GEOJSON=data/countries.geojson \
    MIGRATIONS_DIR=migrations \
    UPLOAD_DIR=uploads
EXPOSE 8080
CMD ["/app/server"]
