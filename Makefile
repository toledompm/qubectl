KUBECTL_PLUGINS_DIR = ~/.bin/

.PHONY: build
build:
	@echo "Building..."
	@go build cmd/kubectl-q.go
	@mv kubectl-q $(KUBECTL_PLUGINS_DIR)
