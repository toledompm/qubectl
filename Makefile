KUBECTL_PLUGINS_DIR = ~/.bin/

.PHONY: dev
dev:
	@go mod tidy
	@go mod vendor
	@go vet ./...
	@go fmt ./...

.PHONY: build
build: dev
	@echo "Building..."
	@go build cmd/kubectl-q.go
	@mv kubectl-q $(KUBECTL_PLUGINS_DIR)
