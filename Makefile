REGISTRY ?= docker.io
TAG      ?= 0.17.1
BUILD_TAGS = containers_image_openpgp,exclude_graphdriver_btrfs
IMAGE = $(REGISTRY)/homelab/rsvp.pizza
LDFLAGS := -X 'main.Version=$(TAG)' \
		-X 'main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' \
		-X 'main.GitCommit=$(shell git rev-parse HEAD)'

.PHONY: build image push clean

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build \
		-ldflags "$(LDFLAGS)" \
		-tags "$(BUILD_TAGS)" -o rsvp.pizza

image: build
	podman build --build-arg BINARY=rsvp.pizza \
		--platform linux/amd64 \
		-t $(IMAGE):$(TAG) .

push: image
	podman push $(IMAGE):$(TAG)

clean:
	rm -f rsvp.pizza rsvp.pizza.arm64
