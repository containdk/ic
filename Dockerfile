FROM --platform=${BUILDPLATFORM} golang:alpine@sha256:d3f0cf7723f3429e3f9ed846243970b20a2de7bae6a5b66fc5914e228d831bbb AS base

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
