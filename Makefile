# Setup name variables for the package/tool
NAME := bpfd
PKG := github.com/jessfraz/$(NAME)

CGO_ENABLED := 1

# Set any default go build tags.
BUILDTAGS :=

include basic.mk

.PHONY: prebuild
prebuild:

.PHONY: image-dev
image-dev:
	docker build --rm --force-rm -f Dockerfile.dev -t $(REGISTRY)/$(NAME):dev .

DOCKER_FLAGS+=--rm -i \
	--disable-content-trust=true
DOCKER_FLAGS+=-v $(CURDIR):/go/src/$(PKG)
DOCKER_FLAGS+=--workdir /go/src/$(PKG)

.PHONY: test-container
test-container: image-dev ## Run a command in a test container with all the needed dependencies (ex. CMD=make test).
	@:$(call check_defined, CMD, command to run in the container)
	docker run $(DOCKER_FLAGS) \
		$(REGISTRY)/$(NAME):dev \
		$(CMD)

GRPC_API_DIR=api/grpc
.PHONY:protoc
protoc: $(CURDIR)/$(GRPC_API_DIR)/api.pb.go ## Generate the protobuf files.

$(CURDIR)/$(GRPC_API_DIR)/api.pb.go: image-dev $(CURDIR)/$(GRPC_API_DIR)/api.proto
	@docker run $(DOCKER_FLAGS) \
		$(REGISTRY)/$(NAME):dev \
		protoc -I ./$(GRPC_API_DIR) \
		./$(GRPC_API_DIR)/api.proto \
		--go_out=plugins=grpc:$(GRPC_API_DIR)
