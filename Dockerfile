FROM --platform=${BUILDPLATFORM} golang:1.23 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN go build -v std

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN --mount=type=cache,mode=0777,target=/go/pkg/mod go mod download all

COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./src ./src

RUN --mount=type=cache,mode=0777,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -tags timetzdata -o /server ./cmd/main.go 

# run container
FROM scratch

#Adding root serts for ssl
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /server /app/konfa-server

WORKDIR /app

ENTRYPOINT [ "/app/konfa-server" ]