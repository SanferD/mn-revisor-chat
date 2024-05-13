BOOTSTRAP=bootstrap
BUILD_DIR=.build
CMD_DIR=cmd
LOCALSTACK_IMAGE_NAME="localstack/localstack"
NETWORK_NAME=mn-revisor-chat-network
SETTINGS_FILE_PATH="../settings.env"

define build_lambda
	mkdir -p $(BUILD_DIR)/$(1)
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o $(BUILD_DIR)/$(1)/$(BOOTSTRAP) $(CMD_DIR)/$(1)/$(1).go
	zip -j $(BUILD_DIR)/$(1)/$(1).zip $(BUILD_DIR)/$(1)/$(BOOTSTRAP)
endef

define build_ecs
	mkdir -p $(BUILD_DIR)/$(1)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(1)/$(1) $(CMD_DIR)/$(1)/$(1).go
	$(if $(with_docker),docker build -t $(1) --file ./$(CMD_DIR)/$(1)/Dockerfile --build-arg BINARY_PATH=$(BUILD_DIR)/$(1)/$(1) .)
	# docker build -t $(1) --file ./$(CMD_DIR)/$(1)/Dockerfile --build-arg BINARY_PATH=$(BUILD_DIR)/$(1)/$(1) .
endef

define run_ecs
	docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)
	docker run --network=$(NETWORK_NAME) --name $(1) -d -p 4566:4566 -p 4571:4571 --rm $(LOCALSTACK_IMAGE_NAME)
	docker run --network=$(NETWORK_NAME) --env-file $(SETTINGS_FILE_PATH) --rm $(1)
endef

build-lambda:
	$(call build_lambda,$(cmd))

build-ecs:
	$(call build_ecs,$(cmd),$(build_image))

run-crawler: build-crawler
	$(call run_ecs,crawler)

clean:
	rm -rf $(BUILD_DIR)
