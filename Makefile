.PHONY: all build deps image lint test vet
CHECK_FILES=$$(go list ./... | grep -v /vendor/)
IMAGE_NAME="sylent/code-runner"

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: clean generate lint vet test build ## Run the tests and build the binary.

build: ## Build the binary.
	@mkdir -p "bin"  || true
	go build -C . -o ./bin/code-runner

deps: ## Install dependencies.
	@go install golang.org/x/lint/golint
	@go mod download

generate: ## Generate API Stub from openapi spec
	@ls ~/go/bin | grep "swagger" > /dev/null || (DIR=$$(mktemp -d) && \
		git clone https://github.com/go-swagger/go-swagger $$DIR && \
		cd $$DIR && go install ./cmd/swagger && \
		echo "deleting $$DIR" && \
		rm -rf $$DIR)
	@mkdir -p "gen" || true
	@~/go/bin/swagger generate server -t gen -f ./api/openapi.yml --exclude-main -A cr

image: ## Build the Docker image.
	docker build -t $(IMAGE_NAME) .

push: ## Build the Docker image.
	docker push $(IMAGE_NAME)
lint: ## Lint the code.
	golint $(CHECK_FILES)

test: ## Run tests.
	go test -p 1 -v $(CHECK_FILES)

vet: # Vet the code
	go vet $(CHECK_FILES)

clean: ## Clean binaries and generated sources
	@rm -rf bin/*
	@rm -rf gen/*
