# Copyright (c) karl-cardenas-coding
# SPDX-License-Identifier: MIT

FROM golang:1.23.0-alpine3.20 AS builder

LABEL org.opencontainers.image.source="http://github.com/karl-cardenas-coding/go-lambda-cleanup"
LABEL org.opencontainers.image.description="A solution for removing previous versions of AWS Lambdas"

ARG VERSION

ADD ./ /source
RUN cd /source && \
adduser -H -u 1002 -D appuser appuser && \
go build -ldflags="-X 'github.com/karl-cardenas-coding/go-lambda-cleanup/v2/cmd.VersionString=${VERSION}'" -o glc -v

FROM alpine:latest

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder --chown=appuser:appuser  /source/glc /usr/bin/

RUN apk -U upgrade --no-cache
USER appuser

ENTRYPOINT ["/usr/bin/glc"]

