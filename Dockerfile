FROM golang:1.18.2-alpine3.15 as builder

LABEL org.opencontainers.image.source="http://github.com/karl-cardenas-coding/go-lambda-cleanup"

ARG VERSION
ARG OS linux
ARG ARCH amd64

RUN 
ADD ./ /source
RUN cd /source && \
GOOS=$OS GOARCH=$ARCH go build -ldflags="-X 'github.com/karl-cardenas-coding/go-lambda-cleanup/cmd.VersionString=$VERSION'" -o glc -v && \
adduser -H -u 1002 -D appuser appuser


FROM alpine:latest

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder --chown=appuser:appuser  /source/glc /usr/bin/

RUN apk -U upgrade --no-cache
USER appuser

ENTRYPOINT ["/usr/bin/glc"]

