# CI-Image fuer ci/gate.sh: alle Werkzeuge der Gates, versionsgepinnt.
#
# Der Tag wird aus dieser Datei gebildet (ci/deps-tag.sh). Eine neue
# Werkzeugversion ist eine Aenderung hier und damit ein neues Image; die
# Pipeline baut es beim ersten Lauf, alle weiteren Laeufe ziehen es nur.

FROM golang:1.25.0-bookworm

ARG STATICCHECK_VERSION=2025.1.1
ARG GOSEC_VERSION=v2.22.8
ARG GOVULNCHECK_VERSION=v1.1.4
ARG ACTIONLINT_VERSION=v1.7.7
ARG NODE_MAJOR=22
ARG PNPM_VERSION=11.29.0

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates curl git python3 python3-yaml shellcheck build-essential \
    && curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && npm install -g "pnpm@${PNPM_VERSION}" \
    && rm -rf /var/lib/apt/lists/*

ENV GOBIN=/usr/local/bin GOTOOLCHAIN=local
RUN go install "honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION}" \
    && go install "github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}" \
    && go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" \
    && go install "github.com/rhysd/actionlint/cmd/actionlint@${ACTIONLINT_VERSION}" \
    && rm -rf /root/.cache/go-build /go/pkg/mod

WORKDIR /work
