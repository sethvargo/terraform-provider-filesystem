dev:
	@go install ./...
.PHONY: dev

generate:
	@rm -rf docs/
	@go generate ./...
.PHONY: generate

test:
	@go test -count=1 -shuffle=on -short ./...
.PHONY: test

test-acc:
	@TF_ACC=1 go test -count=1 -shuffle=on -race ./...
.PHONY: test-acc
