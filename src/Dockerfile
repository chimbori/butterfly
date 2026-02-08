# Build Go binary

FROM golang:1.25.7 AS builder
RUN go version

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN set -Eeux && \
  go mod download && \
  go mod verify
COPY . .

RUN go build \
  -trimpath \
  -ldflags="-w -s -X 'butterfly.chimbori.dev/conf.BuildTimestamp=$(TZ=":America/Los_Angeles" date +"%Y-%m-%d %H:%M:%S")'"

# Copy to deployable image, with Chrome Headless

FROM chromedp/headless-shell:latest
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /app/butterfly /butterfly
RUN useradd -r -s /usr/sbin/nologin butterfly && \
  mkdir -p /data && \
  chown -R butterfly:butterfly /data /butterfly
EXPOSE 9999

ENV TZ="America/Los_Angeles"
USER butterfly
ENTRYPOINT ["/butterfly"]
CMD ["--config=/data/butterfly.yml"]
