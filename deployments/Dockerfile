
FROM golang:1.25-alpine AS builder


RUN apk add --no-cache ca-certificates git tzdata wget

WORKDIR /build


COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w -extldflags '-static'" \
      -trimpath \
      -o /bin/gateway \
      ./cmd/gateway



FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /usr/bin/wget /usr/bin/wget

COPY --from=builder /bin/gateway /gateway

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/gateway"]
