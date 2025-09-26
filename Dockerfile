# syntax=docker/dockerfile:1

FROM mcr.microsoft.com/devcontainers/go:bookworm AS dev

USER vscode

RUN . /etc/os-release && \
    curl -fsSL https://packages.microsoft.com/config/debian/$VERSION_ID/packages-microsoft-prod.deb > ms-packages.deb \
    && sudo dpkg -i ms-packages.deb \
    && sudo apt-get update \
    && sudo apt-get install -y --no-install-recommends powershell-lts \
    && sudo rm -rf /var/lib/apt/lists/* \
    && rm ms-packages.deb

RUN go install github.com/a-h/templ/cmd/templ@v0.3.943

RUN curl -fsSL https://github.com/cli/cli/releases/download/v2.79.0/gh_2.79.0_linux_amd64.deb > gh.deb \
    && sudo dpkg -i gh.deb \
    && rm gh.deb

RUN --mount=type=secret,id=build-secrets,uid=1000 \
    set -o allexport \
    && . /run/secrets/build-secrets \
    && gh extension install github/gh-gei

FROM golang:latest AS build

WORKDIR /src

COPY src/ghec-migrator/go.mod src/ghec-migrator/go.sum .
RUN go mod download

COPY src/ghec-migrator/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ghec-migrator

FROM registry.access.redhat.com/ubi9/ubi AS release

ENV HOME=/home/default

RUN useradd -u 1001 -r -g 0 -d ${HOME} -c "Default Application User" default

RUN dnf install -y https://github.com/PowerShell/PowerShell/releases/download/v7.5.3/powershell-7.5.3-1.rh.x86_64.rpm \
    && dnf clean all

COPY --from=dev /usr/bin/gh /usr/bin/gh
COPY --from=dev /home/vscode/.local/share/gh/extensions/gh-gei ${HOME}/.local/share/gh/extensions/gh-gei
COPY --from=build /ghec-migrator /ghec-migrator

EXPOSE 8080

USER default

ENTRYPOINT ["/ghec-migrator"]
