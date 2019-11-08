NAME=terraform-provider-filesystem
GOOSES = darwin linux windows
GOARCHES = amd64

deps:
	@go get -u ./...
	@go mod tidy
.PHONY: deps

dev:
	@go install ./...
.PHONY: dev

build:
	@rm -rf build/
	@for GOOS in ${GOOSES}; do \
		for GOARCH in ${GOARCHES}; do \
			echo "Building $${GOOS}/$${GOARCH}" ; \
			go build \
				-a \
				-ldflags "-s -w -extldflags 'static'" \
				-installsuffix cgo \
				-tags netgo \
				-o build/${NAME}_$${GOOS}_$${GOARCH} \
				. ; \
		done ; \
	done
.PHONY: build

test:
	@go test -short -parallel=40 ./...
.PHONY: test

test-acc:
	@go test -parallel=40 -count=1 ./...
.PHONY: test-acc
