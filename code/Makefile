BUILD_DIR=.build
BOOTSTRAP=bootstrap

define build_target
	mkdir -p $(BUILD_DIR)/$(1)
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o $(BUILD_DIR)/$(1)/$(BOOTSTRAP) aws/$(1)/$(1).go
	zip -j $(BUILD_DIR)/$(1)/$(1).zip $(BUILD_DIR)/$(1)/$(BOOTSTRAP)
endef

build: build-trigger_crawler build-crawler

build-trigger_crawler:
	$(call build_target,trigger_crawler)

build-crawler:
	$(call build_target,crawler)

clean:
	rm -rf $(BUILD_DIR)
