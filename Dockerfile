FROM golang:1.24-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG MAIN_PKG=.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ${MAIN_PKG}

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/app /app/app

COPY --from=build --chown=65532:65532 /src/config /app/config

COPY --from=build --chown=65532:65532 /src/cert /app/cert
COPY --from=build --chown=65532:65532 /src/cmd/changelog /app/cmd/changelog

EXPOSE 9000
USER 65532:65532
ENTRYPOINT ["/app/app"]
