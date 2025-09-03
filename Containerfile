FROM golang:alpine3.22 AS builder

# Install git for go package dependencies
RUN apk update && apk add --no-cache git

# Create appuser
ENV USER=snmpipe
ENV UID=10001
ENV BINDIR=/usr/local/bin
ENV CONFDIR=/etc/snmpipe

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid $UID \
    $USER

# Change workdir and copy files
WORKDIR $GOPATH/src/snmpipe/
COPY main.go .
COPY helper.go .
COPY poll.go .
COPY trap.go .
COPY splunk.go .

# Fetch dependencies and build the binary
RUN go mod init snmpipe && go mod tidy
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o $BINDIR/snmpipe
RUN mkdir $CONFDIR && chown $USER:$USER $BINDIR/snmpipe && chown $USER:$USER $CONFDIR

FROM scratch

LABEL org.opencontainers.image.title="snmpipe-debug" \
      org.opencontainers.image.description="SNMP poller and trap receiver that sends data to Splunk HEC" \
      org.opencontainers.image.version="0.2.0" \
      org.opencontainers.image.licenses="GPL-3.0" \
      org.opencontainers.image.authors="Daniel Kofler <fox27374@gmail.com>" \
      org.opencontainers.image.source="https://github.com/fox27374/snmpipe" \
      org.opencontainers.image.documentation="https://github.com/fox27374/snmpipe"

ENV USER=snmpipe
ENV BINDIR=/usr/local/bin
ENV CONFDIR=/etc/snmpipe

# Import the user and group files from builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy executable and config directory
COPY --from=builder $BINDIR/snmpipe $BINDIR/snmpipe
COPY --from=builder $CONFDIR $CONFDIR

# Import other tools from busybox
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Change to unprivileged user
USER $USER:$USER

# Run the snmpipe binary
ENTRYPOINT ["$BINDIR/snmpipe"]
CMD []
