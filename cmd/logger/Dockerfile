FROM golang:1.18 AS build
WORKDIR /go/src/
COPY . /go/src/
ENV CGO_ENABLED=0
ENV GOPATH=/go
ARG VERSION=dev

WORKDIR /go/src/cmd/bogie-edge
RUN ls
RUN go mod tidy && GOOS=linux go build \
    -ldflags "-X github.com/ci4rail/bogie-pdm/internal/version.Version=${VERSION}" \
    -o /install/bogie-edge main.go

FROM alpine:3.12
COPY --from=build /install/bogie-edge /bogie-edge
COPY ./cmd/bogie-edge/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
