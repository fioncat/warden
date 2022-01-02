.PHONY: debug
debug:
	@GO111MODULE=on CGO_ENABLED=0 go build -o output/debug-watch cmd/debug-watch/main.go

.PHONY: build
build:
	@GO111MODULE=on CGO_ENABLED=0 go build -o output/warden cmd/warden/main.go

.PHONY: install
install:
	@cd cmd/warden; GO111MODULE=on CGO_ENABLED=0 go install



