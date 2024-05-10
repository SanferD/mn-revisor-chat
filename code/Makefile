BUILD_DIR=.build
BOOTSTRAP=bootstrap

define build_lambda
	mkdir -p $(BUILD_DIR)/$(1)
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o $(BUILD_DIR)/$(1)/$(BOOTSTRAP) cmd/$(1)/$(1).go
	zip -j $(BUILD_DIR)/$(1)/$(1).zip $(BUILD_DIR)/$(1)/$(BOOTSTRAP)
endef

define build_ecs
	mkdir -p $(BUILD_DIR)/$(1)
	go build -o $(BUILD_DIR)/$(1)/$(1) cmd/$(1)/$(1).go
endef

build: build-trigger_crawler build-crawler

build-trigger_crawler:
	$(call build_lambda,trigger_crawler)

build-crawler:
	$(call build_ecs,crawler)

clean:
	rm -rf $(BUILD_DIR)
