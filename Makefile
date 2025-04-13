.PHONY: build test clean release-major release-minor release-patch release-custom publish publish-latest release-major-push release-minor-push release-patch-push release-custom-push lint lint-fix pre-commit pre-deploy help

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
NEXT_MAJOR_VERSION ?= $(shell echo $(VERSION) | awk -F. '{ printf "v%d.0.0", $$1+1 }')
NEXT_MINOR_VERSION ?= $(shell echo $(VERSION) | awk -F. '{ printf "v%d.%d.0", $$1, $$2+1 }')
NEXT_PATCH_VERSION ?= $(shell echo $(VERSION) | awk -F. '{ printf "v%d.%d.%d", $$1, $$2, $$3+1 }')
COMMIT_MSG ?= "Release $(TAG)"
LATEST_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

build:
	@echo "Building condomngr..."
	@go build -v .

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -f condomngr

# Lint targets
lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.53.3; \
	fi
	@$(shell go env GOPATH)/bin/golangci-lint run ./...

lint-fix:
	@echo "Running linter with auto-fix..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.53.3; \
	fi
	@$(shell go env GOPATH)/bin/golangci-lint run --fix ./...

# Convenience targets for common workflows
pre-commit: lint-fix test
	@echo "Pre-commit checks completed successfully"

pre-deploy: lint test build
	@echo "Pre-deploy checks completed successfully"

# Release targets
release-major: TAG=$(NEXT_MAJOR_VERSION)
release-major: release

release-minor: TAG=$(NEXT_MINOR_VERSION)
release-minor: release

release-patch: TAG=$(NEXT_PATCH_VERSION)
release-patch: release

release-custom:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG is required. Use 'make release-custom TAG=vX.Y.Z'"; \
		exit 1; \
	fi
	@$(MAKE) release TAG=$(TAG)

release:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG is required"; \
		exit 1; \
	fi
	@echo "Creating release $(TAG)..."
	@echo "Current version: $(VERSION)"
	@echo "New version: $(TAG)"
	@git tag -a $(TAG) -m $(COMMIT_MSG)
	@echo "Tag $(TAG) created. To push this release, run:"
	@echo "make publish TAG=$(TAG)"

# One-step release and publish targets
release-major-push: TAG=$(NEXT_MAJOR_VERSION)
release-major-push: release publish

release-minor-push: TAG=$(NEXT_MINOR_VERSION)
release-minor-push: release publish

release-patch-push: TAG=$(NEXT_PATCH_VERSION)
release-patch-push: release publish

release-custom-push:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG is required. Use 'make release-custom-push TAG=vX.Y.Z'"; \
		exit 1; \
	fi
	@$(MAKE) release TAG=$(TAG)
	@$(MAKE) publish TAG=$(TAG)

# Push the tag to trigger a release
publish:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG is required. Use 'make publish TAG=vX.Y.Z'"; \
		exit 1; \
	fi
	@echo "Publishing release $(TAG)..."
	@git push origin $(TAG)
	@echo "Release $(TAG) has been published."
	@echo "GitHub Actions will now build and create the release."

# Push the latest tag
publish-latest:
	@echo "Publishing latest tag $(LATEST_TAG)..."
	@git push origin $(LATEST_TAG)
	@echo "Latest tag $(LATEST_TAG) has been published."
	@echo "GitHub Actions will now build and create the release."

# Display help information
help:
	@echo "Available targets:"
	@echo "  build               - Build the application"
	@echo "  test                - Run tests"
	@echo "  clean               - Remove build artifacts"
	@echo "  lint                - Run linter to check for issues"
	@echo "  lint-fix            - Run linter and automatically fix common issues"
	@echo "  pre-commit          - Run lint-fix and tests (use before committing)"
	@echo "  pre-deploy          - Run lint, tests, and build (use before deploying)"
	@echo ""
	@echo "  # Two-step release process:"
	@echo "  release-major       - Create a new major release tag (vX.0.0)"
	@echo "  release-minor       - Create a new minor release tag (vX.Y.0)"
	@echo "  release-patch       - Create a new patch release tag (vX.Y.Z)"
	@echo "  release-custom      - Create a custom release tag (make release-custom TAG=vX.Y.Z)"
	@echo "  publish             - Push a tag to trigger the release process (make publish TAG=vX.Y.Z)"
	@echo "  publish-latest      - Push the latest tag to trigger the release process"
	@echo ""
	@echo "  # One-step release process:"
	@echo "  release-major-push  - Create and publish a new major release"
	@echo "  release-minor-push  - Create and publish a new minor release"
	@echo "  release-patch-push  - Create and publish a new patch release"
	@echo "  release-custom-push - Create and publish a custom release (make release-custom-push TAG=vX.Y.Z)"
	@echo ""
	@echo "Current version: $(VERSION)"
	@echo "Next major: $(NEXT_MAJOR_VERSION)"
	@echo "Next minor: $(NEXT_MINOR_VERSION)"
	@echo "Next patch: $(NEXT_PATCH_VERSION)"

# Default target
.DEFAULT_GOAL := help 