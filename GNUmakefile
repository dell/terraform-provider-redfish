PKG_NAME=redfish
VERSION=1.1.0
TEST?=$$(go list ./... | grep -v 'vendor')
INSTALL_ROOT?=~/.terraform.d/plugins
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
HOSTNAME=registry.terraform.io
NAMESPACE=dell
BINARY=terraform-provider-${PKG_NAME}
OS_ARCH=linux_amd64

default: build

build:
	go install
	go build -o $(CURDIR)/bin/${OS_ARCH}/${BINARY}_v$(VERSION)

# formats all .go files
fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w $(GOFMT_FILES)

# runs a Go format check
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

check:
	golangci-lint run --fix

lint:
	@echo "==> Checking source code against linters..."
	# TODO: renable - tfproviderlint ./redfish
	golangci-lint run --fix

# vets all .go files
vet:
	@echo "go vet ."
	@go vet $$(go list ./... ) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

gosec:
	gosec -exclude-generated ./...

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p $(INSTALL_ROOT)/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}
	mv $(CURDIR)/bin/${OS_ARCH}/${BINARY}_v$(VERSION) $(INSTALL_ROOT)/${HOSTNAME}/${NAMESPACE}/${PKG_NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

generate:
	GOFLAGS="-mod=readonly" go generate ./...

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-local PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-lint:
	@echo "==> Checking website against linters..."
	@misspell -error -source=text website/

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

clean:
	go clean --cache
	rm -rf vendor bin

.PHONY: build test testacc vet fmt fmtcheck errcheck lint tools test-compile website website-lint website-test
