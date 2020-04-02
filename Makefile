export GOPROXY=direct

.PHONY: all
all:
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  clean          - remove bin, vendor directories"
	@echo "  clean-install  - clean and then install"
	@echo "  install        - install vendored dependencies"
	@echo "  build          - build production binaries"
	@echo "  lint           - lint and vet code"
	@echo "  run            - runs the API server"
	@echo "  test           - runs tests"

.PHONY: clean
clean:
	@rm -rf vendor

.PHONY: clean-install
clean-install: clean install

.PHONY: install
install:
	@go mod vendor

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/wm ./cmd/wm

.PHONY: lint
lint:
	@go vet $$(go list ./...)
	@go list ./... | grep -v /vendor/ | xargs -L1 golint --set_exit_status

.PHONY: run
run:
	@go run cmd/wm/main.go

.PHONY: test
test:
	@go test -race -cover $$(go list ./...)
