FROM golang:alpine AS builder

# Install git for go package dependencies
RUN apk update && apk add --no-cache git

# Create appuser
ENV USER=appuser
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
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/snmpipe

# Create new container and copy binary file
# as well as the group and passwd file
FROM scratch

# Import the user and group files from the builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy executable and config file
COPY --from=builder /go/bin/snmpipe /go/bin/snmpipe
COPY config.json .

# Change to unprivileged user
USER appuser:appuser

# Run the snmpipe binary
ENTRYPOINT ["/go/bin/snmpipe"]