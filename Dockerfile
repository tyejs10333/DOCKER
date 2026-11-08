FROM golang:1.27.3-alpine AS build
ENV GOTOOLCHAIN=local CGO_ENABLED=0
WORKDIR /src
COPY ["Stopka Backend/go.mod", "Stopka Backend/go.sum", "./"]
RUN go mod download
COPY ["Stopka Backend/", "./"]
RUN go build -trimpath -ldflags="-s -w" -o /out/stopka ./cmd/stopka

FROM alpine:3
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S stopka \
    && adduser -S -G stopka -H -s /sbin/nologin stopka
WORKDIR /app
COPY --from=build /out/stopka /app/stopka
COPY ["Stopka Backend/migrations/", "/app/migrations/"]
COPY ["Stopka Frontend/", "/app/frontend/"]
ENV HTTP_ADDR=:8080 STATIC_DIR=/app/frontend
USER stopka
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
CMD ["/app/stopka"]