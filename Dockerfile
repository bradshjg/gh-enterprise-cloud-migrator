# syntax=docker/dockerfile:1

FROM mcr.microsoft.com/devcontainers/go AS dev

USER vscode

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

COPY src/ghes-to-ghec/go.mod src/ghes-to-ghec/go.sum .
RUN go mod download

COPY src/ghes-to-ghec/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ghes-to-ghec

FROM registry.access.redhat.com/ubi9/ubi AS release

ENV HOME=/home/default

RUN useradd -u 1001 -r -g 0 -d ${HOME} -c "Default Application User" default

RUN INSTALL_PKGS="libicu" \
    && yum install -y --setopt=tsflags=nodocs $INSTALL_PKGS \
    && rpm -V $INSTALL_PKGS \
    && yum -y clean all --enablerepo='*'

COPY --from=dev /usr/bin/gh /usr/bin/gh
COPY --from=dev /home/vscode/.local/share/gh/extensions/gh-gei ${HOME}/.local/share/gh/extensions/gh-gei
COPY --from=build /ghes-to-ghec /ghes-to-ghec

EXPOSE 8080

USER default

ENTRYPOINT ["/ghes-to-ghec"]
