VERSION 0.7

ARG --global TARGET_DOCKER_REGISTRY=ghcr.io/librepod

FROM golang:1.18-bullseye

validate-pr:
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/frp-port-keeper ./main.go
  SAVE ARTIFACT build /build AS LOCAL ./build

multi-build:
  ARG RELEASE_VERSION=latest
  ENV PLATFORMS="darwin/amd64 darwin/arm64 windows/amd64 linux/amd64 linux/arm64"
  ENV VERSION_INJECT="github.com/librepod/frp-port-keeper/main.Version"
  ENV OUTPUT_PATH_FORMAT="./build/${RELEASE_VERSION}/{{.OS}}/{{.Arch}}/frp-port-keeper"

  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN go install github.com/mitchellh/gox@v1.0.1 \
      && gox -osarch="${PLATFORMS}" -ldflags "-X ${VERSION_INJECT}=${RELEASE_VERSION}" -output "${OUTPUT_PATH_FORMAT}"

  SAVE ARTIFACT build /build AS LOCAL ./build

build:
  ARG RELEASE_VERSION=latest
  ENV GOOS=linux
  ENV GOARCH=amd64

  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN go build
  SAVE ARTIFACT main /frp-port-keeper AS LOCAL ./frp-port-keeper

release:
  FROM  +multi-build

  ARG --required GITHUB_TOKEN
  ARG --required RELEASE_VERSION
  ARG OUT_BASE="./build/${RELEASE_VERSION}"

  # Install gh-cli
  RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
      && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
      && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
      && apt-get update && apt-get install gh jq -y \
      && gh --version

  # Generate release notes
  RUN gh api -X POST 'repos/librepod/frp-port-keeper/releases/generate-notes' \
        -F commitish=${RELEASE_VERSION} \
        -F tag_name=${RELEASE_VERSION} \
      > tmp-release-notes.json \
      && cat ./tmp-release-notes.json | jq

  # Gzip the bins
  RUN tar -czvf "${OUT_BASE}/darwin/amd64/frp-port-keeper_darwin_amd64.tar.gz" -C "${OUT_BASE}/darwin/amd64" frp-port-keeper \
      && tar -czvf "${OUT_BASE}/darwin/arm64/frp-port-keeper_darwin_arm64.tar.gz" -C "${OUT_BASE}/darwin/arm64" frp-port-keeper \
      && tar -czvf "${OUT_BASE}/windows/amd64/frp-port-keeper_windows_amd64.tar.gz" -C "${OUT_BASE}/windows/amd64" frp-port-keeper.exe \
      && tar -czvf "${OUT_BASE}/linux/amd64/frp-port-keeper_linux_amd64.tar.gz" -C "${OUT_BASE}/linux/amd64" frp-port-keeper \
      && tar -czvf "${OUT_BASE}/linux/arm64/frp-port-keeper_linux_arm64.tar.gz" -C "${OUT_BASE}/linux/arm64" frp-port-keeper

  # Create release
  RUN jq -r .body tmp-release-notes.json > tmp-release-notes.md \
      && gh release create ${RELEASE_VERSION} \
        --title "$(jq -r .name tmp-release-notes.json)" \
        --notes-file tmp-release-notes.md \
        --verify-tag \
        --draft \
        "${OUT_BASE}/darwin/amd64/frp-port-keeper_darwin_amd64.tar.gz#frp-port-keeper_osx_amd64" \
        "${OUT_BASE}/darwin/arm64/frp-port-keeper_darwin_arm64.tar.gz#frp-port-keeper_osx_arm64" \
        "${OUT_BASE}/windows/amd64/frp-port-keeper_windows_amd64.tar.gz#frp-port-keeper_windows_amd64" \
        "${OUT_BASE}/linux/amd64/frp-port-keeper_linux_amd64.tar.gz#frp-port-keeper_linux_amd64" \
        "${OUT_BASE}/linux/arm64/frp-port-keeper_linux_arm64.tar.gz#frp-port-keeper_linux_arm64"
  
  SAVE ARTIFACT build /build AS LOCAL ./build

image:
  # TODO: the built binary doesn't work in alpine images
  # FROM golang:1.18-alpine3.17
  FROM golang:1.18-bullseye
  LABEL org.opencontainers.image.source="https://github.com/librepod/frp-port-keeper"
  ARG user=gopher
  ARG group=gopher
  ARG uid=1000
  ARG gid=1000
  ARG HOME=/app

  RUN mkdir -p ${HOME}  \
      && groupadd -g ${gid} ${group}  \
      && useradd --home-dir ${HOME} -u ${uid} -g ${group} -s /bin/bash ${user}

  ARG RELEASE_VERSION=latest
  WORKDIR ${HOME}
  COPY +build/frp-port-keeper ./frp-port-keeper
  EXPOSE 8080
  CMD ["/app/frp-port-keeper"]
  SAVE IMAGE --push ${TARGET_DOCKER_REGISTRY}/frp-port-keeper:$RELEASE_VERSION

validate-mr:
  # Smoke test the application
  # TODO: set proper validation
  BUILD +build

