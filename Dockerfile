FROM golang:1.19-alpine as build

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o tplinkexporter-plus .

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=build /build/tplinkexporter-plus /tplinkexporter-plus
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Add labels for better container management
LABEL org.opencontainers.image.title="TP-Link Switch Exporter - PLUS"
LABEL org.opencontainers.image.description="Enhanced Prometheus exporter for TP-Link EasySmart switches"
LABEL org.opencontainers.image.vendor="thelastguardian"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/thelastguardian/tplinkexporter-plus"

# Create a non-root user (even though we're using scratch, this documents intent)
USER 65534:65534

EXPOSE 9717

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/tplinkexporter-plus", "--help"]

ENTRYPOINT ["/tplinkexporter-plus"]