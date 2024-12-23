FROM golang:1.23 AS build
RUN go build -v std

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./src ./src

RUN go build -tags timetzdata -o /server ./cmd/main.go 

# run container
FROM scratch

#Adding root serts for ssl
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /server /app/konfa-server

WORKDIR /app

ENTRYPOINT [ "/app/konfa-server" ]