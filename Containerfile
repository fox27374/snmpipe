FROM golang:alpine3.22 AS builder

# Install git for go package dependencies
RUN apk update && apk add --no-cache git

# Create appuser
ENV USER=snmpipe
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Change workdir and copy files
WORKDIR $GOPATH/src/snmpipe/
COPY main.go .
COPY helper.go .
COPY poll.go .
COPY trap.go .
COPY splunk.go .

# Fetch dependencies and build the binary
RUN go mod init snmpipe && go mod tidy
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/snmpipe

# Create new container and copy binary file
# as well as the group and passwd file
FROM scratch

LABEL org.opencontainers.image.title="snmpipe" \
      org.opencontainers.image.description="SNMP poller and trap receiver that sends data to Splunk HEC" \
      org.opencontainers.image.version="0.2.0" \
      org.opencontainers.image.licenses="GPL-3.0" \
      org.opencontainers.image.authors="Daniel Kofler <fox27374@gmail.com>" \
      org.opencontainers.image.source="https://github.com/fox27374/snmpipe" \
      org.opencontainers.image.documentation="https://github.com/fox27374/snmpipe"

ENV USER=snmpipe

# Import the user and group files from the builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy executable
COPY --from=builder --chown=${USER}:${USER} /app/ /app/

WORKDIR /app

# Change to unprivileged user
USER ${USER}:${USER}

# Run the snmpipe binary
ENTRYPOINT ["snmpipe"]