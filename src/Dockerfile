# Build Go binary

FROM golang:1.26.5 AS builder
RUN go version

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN set -Eeux && \
  go mod download && \
  go mod verify
COPY . .

RUN go build \
  -o wynglet \
  -trimpath \
  -ldflags="-w -s -X 'wynglet.chimbori.dev/conf.BuildTimestamp=$(TZ="America/Los_Angeles" date +"%Y-%m-%d %H:%M:%S")'"

# Copy to deployable image, with Chrome Headless

FROM chromedp/headless-shell:latest
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /app/wynglet /wynglet

# Install tini to reap zombie Chrome processes and forward signals to the child process.
RUN apt-get update && apt-get install -y tini && rm -rf /var/lib/apt/lists/*

RUN useradd -r -s /usr/sbin/nologin wynglet && \
  mkdir -p /data && \
  chown -R wynglet:wynglet /data /wynglet
EXPOSE 9999

ENV TZ="America/Los_Angeles"
USER wynglet

# Use tini as PID 1 to handle signal forwarding and zombie reaping.
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/wynglet", "--config=/data/wynglet.yml"]
