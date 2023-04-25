PROJECT          := github.com/pulumi/kubespy
TESTPARALLELISM := 10

ensure::
	go mod tidy

build::
	go build $(PROJECT)

lint::
	golangci-lint run

install::
	go install $(PROJECT)

test_all::
	go test -v -cover -timeout 1h -parallel ${TESTPARALLELISM} ./...

