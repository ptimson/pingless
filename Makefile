# ---- config ----
IMAGE ?= ptimson/pingless
VERSION ?= 1.0.0
PLATFORMS ?= linux/amd64,linux/arm64
BUILDER ?= xbuilder

# ---- helpers ----
.PHONY: help show ensure-builder localbuild build release run clean
help:
	@echo "make localbuild       # build local binary to ./bin/pingless"
	@echo "make build            # build local arch build for local testing"
	@echo "make release          # buildx multi-arch build and push"
	@echo "make run              # run container with .env and NET_RAW cap"
	@echo "make clean            # remove local image:$(VERSION)"

show:
	@echo "IMAGE     = $(IMAGE)"
	@echo "VERSION   = $(VERSION)"
	@echo "PLATFORMS = $(PLATFORMS)"
	@echo "BUILDER   = $(BUILDER)"

# Create/select a buildx builder once
ensure-builder:
	@docker buildx inspect $(BUILDER) >/dev/null 2>&1 || docker buildx create --name $(BUILDER) --use
	@docker buildx use $(BUILDER)
	@docker buildx inspect --bootstrap >/dev/null

# ---- builds ----
localbuild:
	@mkdir -p bin
	go build -o bin/pingless ./main.go

# local build for testing
build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):latest \
		.

# multi-arch release to Docker Hub
release: ensure-builder
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):latest \
		--push \
		.

# ---- convenience ----
run:
	docker run --rm -it \
		--env-file .env \
		--cap-add=NET_RAW \
		--name pingless \
		$(IMAGE):$(VERSION)

clean:
	-docker rmi $(IMAGE):$(VERSION) $(IMAGE):latest 2>/dev/null || true
