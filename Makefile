PROJECT          := github.com/pulumi/kubespy

VERSION         := $(shell scripts/get-version)

VERSION_FLAGS   := -ldflags "-X github.com/pulumi/kubespy/version.Version=${VERSION}"

GO              ?= go
GOMETALINTERBIN ?= gometalinter
GOMETALINTER    :=${GOMETALINTERBIN} --config=Gometalinter.json

TESTPARALLELISM := 10
TESTABLE_PKGS   := ./...

build::
	$(GO) build $(VERSION_FLAGS) $(PROJECT)

rel-darwin::
	GOOS=darwin GOARCH=386 $(GO) build -o releases/kubespy-darwin-386/kubespy $(VERSION_FLAGS) $(PROJECT)
	tar -zcvf releases/kubespy-darwin-368.tar.gz releases/kubespy-darwin-386
	GOOS=darwin GOARCH=amd64 $(GO) build -o releases/kubespy-darwin-amd64/kubespy $(VERSION_FLAGS) $(PROJECT)
	tar -zcvf releases/kubespy-darwin-amd64.tar.gz releases/kubespy-darwin-amd64

rel-linux::
	GOOS=linux GOARCH=386 $(GO) build -o releases/kubespy-linux-386/kubespy $(VERSION_FLAGS) $(PROJECT)
	tar -zcvf releases/kubespy-linux-386.tar.gz releases/kubespy-linux-386
	GOOS=linux GOARCH=amd64 $(GO) build -o releases/kubespy-linux-amd64/kubespy $(VERSION_FLAGS) $(PROJECT)
	tar -zcvf releases/kubespy-linux-amd64.tar.gz releases/kubespy-linux-amd64

rel-windows::
	GOOS=windows GOARCH=386 $(GO) build -o releases/kubespy-windows-386/kubespy $(VERSION_FLAGS) $(PROJECT)
	tar -zcvf releases/kubespy-windows-386.tar.gz releases/kubespy-windows-386
	GOOS=windows GOARCH=amd64 $(GO) build -o releases/kubespy-windows-amd64/kubespy $(VERSION_FLAGS) $(PROJECT)
	tar -zcvf releases/kubespy-windows-amd64.tar.gz releases/kubespy-windows-amd64

lint::
	$(GOMETALINTER) ./... | sort ; exit "$${PIPESTATUS[0]}"

install::
	$(GO) install $(VERSION_FLAGS) $(PROJECT)

test_all:: test_fast
	$(GO) test -v -cover -timeout 1h -parallel ${TESTPARALLELISM} $(TESTABLE_PKGS)

.PHONY: check_clean_worktree
check_clean_worktree:
	$$(go env GOPATH)/src/github.com/pulumi/scripts/ci/check-worktree-is-clean.sh

# The travis_* targets are entrypoints for CI.
.PHONY: travis_cron travis_push travis_pull_request travis_api
travis_cron: all
travis_push: only_build check_clean_worktree publish_tgz only_test publish_packages
travis_pull_request: all check_clean_worktree
travis_api: all
