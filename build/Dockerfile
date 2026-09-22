FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH VERSION=dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X main.version=$VERSION" -o /out/kmind-agent-docker ./cmd/kmind-agent-docker

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/kmind-agent-docker /kmind-agent-docker
USER 65532:65532
ENTRYPOINT ["/kmind-agent-docker"]
