.PHONY: build install test test-coverage testacc fmt docs clean test-e2e

# Default target
.DEFAULT_GOAL := help

# Build the provider
build:
	@echo "==> Building the provider..."
	go build -buildvcs=false -o terraform-provider-utilities

# Install the provider
install: build
	@echo "==> Installing the provider..."
	go install

# Install provider locally for Terraform to use
install-local: build
	@echo "==> Installing provider locally for Terraform..."
	@VERSION="0.1.0" \
	PLATFORM="linux_amd64" \
	PLUGIN_DIR_REGISTRY="$$HOME/.terraform.d/plugins/registry.terraform.io/tfstack/utilities/$$VERSION/$$PLATFORM" \
	PLUGIN_DIR_SOURCE="$$HOME/.terraform.d/plugins/hashicorp.com/tfstack/utilities/$$VERSION/$$PLATFORM" \
	&& mkdir -p "$$PLUGIN_DIR_REGISTRY" "$$PLUGIN_DIR_SOURCE" \
	&& cp terraform-provider-utilities "$$PLUGIN_DIR_REGISTRY/" \
	&& cp terraform-provider-utilities "$$PLUGIN_DIR_SOURCE/" \
	&& echo "✅ Provider installed to $$PLUGIN_DIR_REGISTRY" \
	&& echo "✅ Provider installed to $$PLUGIN_DIR_SOURCE"

# Setup .terraformrc for local development
setup-terraformrc: install-local
	@echo "==> Setting up .terraformrc for local development..."
	@mkdir -p $$HOME/.terraform.d/plugins/hashicorp.com/tfstack/utilities/0.1.0/linux_amd64
	@echo 'provider_installation {' > $$HOME/.terraformrc
	@echo '  dev_overrides {' >> $$HOME/.terraformrc
	@echo '    "hashicorp.com/tfstack/utilities" = "$$HOME/.terraform.d/plugins/hashicorp.com/tfstack/utilities/0.1.0/linux_amd64"' >> $$HOME/.terraformrc
	@echo '  }' >> $$HOME/.terraformrc
	@echo '  direct {}' >> $$HOME/.terraformrc
	@echo '}' >> $$HOME/.terraformrc
	@echo "✅ .terraformrc configured"
	@echo ""
	@echo "⚠️  IMPORTANT: Skip 'terraform init' when using dev_overrides!"
	@echo "   Use 'terraform plan' or 'terraform apply' directly."

# Run tests
test:
	@echo "==> Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "==> Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	@echo "==> Coverage report:"
	@go tool cover -func=coverage.out
	@echo ""
	@echo "==> HTML coverage report generated: coverage.html"
	@go tool cover -html=coverage.out -o coverage.html

# Run acceptance tests
testacc:
	@echo "==> Running acceptance tests..."
	TF_ACC=1 go test -v ./...

# Format code
fmt:
	@echo "==> Formatting code..."
	go fmt ./...
	terraform fmt -recursive ./examples/

# Generate documentation
docs:
	@echo "==> Generating documentation..."
	@cd tools && GOFLAGS=-buildvcs=false go generate ./...

# Initialize Terraform in all examples (skip init when using dev_overrides)
init-examples: setup-terraformrc
	@echo "==> Examples ready (skip terraform init when using dev_overrides)"
	@echo "   Use: cd examples/data-sources/utilities_bcrypt_hash && terraform plan"

# End-to-end test for bcrypt hash
test-e2e: setup-terraformrc
	@echo "==> Running end-to-end test for bcrypt hash..."
	@cd examples/data-sources/utilities_bcrypt_hash && \
	rm -rf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup 2>/dev/null || true && \
	terraform plan -out=tfplan && \
	terraform apply tfplan && \
	terraform show -json | grep -q '"hash"' && echo "✅ E2E test passed" || (echo "❌ E2E test failed" && exit 1)

# Initialize a specific example
init-example: install-local
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "Usage: make init-example EXAMPLE=examples/data-sources/utilities_channels"; \
		exit 1; \
	 fi
	@echo "==> Initializing $(EXAMPLE)..."
	@cd $(EXAMPLE) && terraform init -upgrade

# Clean build artifacts
clean:
	@echo "==> Cleaning..."
	rm -f terraform-provider-utilities
	rm -f terraform-provider-utilities.exe
	rm -f coverage.out coverage.html
	go clean
	@echo "==> Cleaning Terraform state files..."
	@find examples -name ".terraform" -type d -exec rm -rf {} + 2>/dev/null || true
	@find examples -name ".terraform.lock.hcl" -type f -delete 2>/dev/null || true
	@find examples -name "*.tfstate" -type f -delete 2>/dev/null || true
	@find examples -name "*.tfstate.*" -type f -delete 2>/dev/null || true

# Help target
help:
	@echo "Available targets:"
	@echo "  build            - Build the provider binary"
	@echo "  install          - Install the provider to GOPATH/bin"
	@echo "  install-local    - Install provider locally for Terraform testing"
	@echo "  setup-terraformrc - Install provider and configure .terraformrc for local dev"
	@echo "  init-examples    - Examples ready (skip terraform init with dev_overrides)"
	@echo "  init-example     - Initialize a specific example (use EXAMPLE=path)"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  testacc          - Run acceptance tests (requires TF_ACC=1)"
	@echo "  test-e2e         - Run end-to-end validation test for bcrypt hash"
	@echo "  fmt              - Format code"
	@echo "  docs             - Generate documentation"
	@echo "  clean            - Clean build artifacts and Terraform state"
	@echo "  help             - Show this help message"
