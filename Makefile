all: container
VERSION=a7
GO_LDFLAGS = -tags 'netgo osusergo static_build' -ldflags "-X github.com/ci4rail/bogie-pdm/internal/version.Version=${VERSION}"

container:
	docker buildx build --build-arg VERSION=${VERSION} -f cmd/bogie-edge/Dockerfile \
		--platform linux/arm64 --push -t ci4rail/bogie-edge:${VERSION} .

bogie-edge-static:
	GOOS=linux GOARCH=arm64 go build $(GO_LDFLAGS) -o ./bin/bogie-edge-static ./cmd/bogie-edge/main.go


.PHONY: all container bogie-edge-static
