PROJECT          := github.com/pulumi/kubespy

VERSION         := $(shell scripts/get-version)

VERSION_FLAGS   := -ldflags "-X github.com/pulumi/kubespy/version.Version=${VERSION}"

TESTPARALLELISM := 10

ensure::
	go mod tidy

build::
	go build $(VERSION_FLAGS) $(PROJECT)

lint::
	golangci-lint run

install::
	go install $(VERSION_FLAGS) $(PROJECT)

test_all::
	go test -v -cover -timeout 1h -parallel ${TESTPARALLELISM} ./...

