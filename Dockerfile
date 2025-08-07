FROM --platform=${BUILDPLATFORM} golang:alpine@sha256:c8c5f95d64aa79b6547f3b626eb84b16a7ce18a139e3e9ca19a8c078b85ba80d AS base

RUN apk update
RUN apk add -U --no-cache ca-certificates && update-ca-certificates
RUN apk add git

RUN adduser -S -u 20000 -H inventory

WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/modx \
    go mod download

FROM base AS builder
ARG TARGETOS
ARG TARGETARCH

ARG VERSION

RUN --mount=target= \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/ic \
    -tags release \
    -ldflags "-s -w -X main.version=${VERSION}"

RUN mkdir /cache
RUN chown 20000 /cache

FROM scratch AS bin-unix
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /out/ic /ic
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 20000

COPY --from=builder --chown=20000 /cache/. /cache/

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM bin-${TARGETOS} AS bin

ENV IC_OIDC_TOKEN_CACHE_DIR=/cache
EXPOSE 18000
ENTRYPOINT ["/ic"]

ARG COMMIT=
ARG VERSION=

LABEL version="$VERSION"
