# kubespy: Kubernetes Resource Observer

kubespy is a Go-based command-line tool for observing Kubernetes resources in real time. It watches and reports information about Kubernetes resources continuously, derived from work done to make Kubernetes deployments predictable in Pulumi's CLI.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build
- **ALWAYS** run dependency management first:
  - `make ensure` -- downloads dependencies (first time ~25 seconds, subsequent runs <1 second). NEVER CANCEL. Set timeout to 5+ minutes for first run.
- **Build the application:**
  - `make build` -- builds the kubespy binary (first time ~1 minute 20 seconds, subsequent builds ~5-6 seconds). NEVER CANCEL. Set timeout to 3+ minutes.
- **Install the application:**
  - `make install` -- installs to $GOPATH/bin, takes ~5 seconds.

### Testing
- **Run tests:**
  - `make test_all` -- runs all tests (takes ~13 seconds first time, <1 second subsequent runs). NEVER CANCEL. Set timeout to 2+ minutes.
  - NOTE: This project has no actual test files (0% coverage), tests complete quickly but don't validate functionality.

### Linting
- **Linting is broken** in this repository due to deprecated linters in .golangci.yml
- `make lint` -- FAILS due to deprecated linters (deadcode, golint, interfacer, structcheck, varcheck) and Go version compatibility issues
- **DO NOT** attempt to fix the linting configuration unless specifically asked to address linting issues
- The CI pipeline (.github/workflows/ci.yml) does NOT run linting, only build and test

### Running kubespy
- **ALWAYS** build the application first with `make build`
- **Basic commands:**
  - `./kubespy --help` -- shows available commands
  - `./kubespy version` -- shows version (returns "dev" for local builds)
  - `./kubespy status --help` -- help for status watching
  - `./kubespy trace --help` -- help for resource tracing
  - `./kubespy changes --help` -- help for change monitoring
  - `./kubespy record --help` -- help for event recording

## Validation

### Manual Testing Requirements
- **ALWAYS** test basic functionality after making changes by running `./kubespy --help` and `./kubespy version`
- **Without Kubernetes cluster:** kubespy will show error "invalid configuration: no configuration has been provided, try setting KUBERNETES_MASTER environment variable" - this is expected and correct behavior
- **Test all subcommands with --help flag** to ensure they display properly

### CI Validation
- **ALWAYS** run `make build` and `make test_all` before committing changes
- The CI (.github/workflows/ci.yml) runs `make build` and `make test_all` using Go 1.24.x on ubuntu-latest
- **DO NOT** run `make lint` as it will fail due to configuration issues

## Common Tasks

### Repository Structure
```
/home/runner/work/kubespy/kubespy/
├── README.md           # Main project documentation
├── CONTRIBUTING.md     # Development guidelines
├── Makefile           # Build automation
├── go.mod             # Go module definition
├── kubespy.go         # Main entry point
├── cmd/               # CLI command implementations
│   ├── root.go        # Root command and parsing
│   ├── status.go      # Status watching command
│   ├── trace.go       # Resource tracing command
│   ├── changes.go     # Change monitoring command
│   ├── record.go      # Event recording command
│   └── version.go     # Version command
├── examples/          # Example usage scenarios
│   ├── trivial-pulumi-example/    # Basic Pod example
│   └── trivial-service-trace-example/ # Service tracing example
├── k8sconfig/         # Kubernetes configuration handling
├── k8sobject/         # Kubernetes object utilities
├── pods/              # Pod-specific functionality
├── print/             # Output formatting
├── version/           # Version information
└── watch/             # Resource watching logic
```

### Key Commands and Expected Times
```bash
# Dependency management (first time: 25 seconds, subsequent: <1 second)
make ensure

# Build process (first time: 1 minute 20 seconds, subsequent: 5-6 seconds) - NEVER CANCEL
make build

# Test suite (first time: 13 seconds, subsequent: <1 second) - NEVER CANCEL  
make test_all

# Install binary (5 seconds)
make install

# Run application
./kubespy --help           # Immediate
./kubespy version          # Immediate, returns "dev"
```

### Known Issues and Workarounds
- **Linting fails:** The .golangci.yml uses deprecated linters. Do not run `make lint` unless fixing linting configuration.
- **No actual tests:** The test suite runs but provides 0% coverage as there are no test files with actual test cases.
- **Requires Kubernetes:** kubespy requires a Kubernetes cluster connection to function. Without it, commands will fail with configuration errors (this is expected).

### Go Version Requirements
- **Go 1.24 or later** required (as specified in go.mod)
- Current CI uses Go 1.24.x

### Dependencies
- Uses Go modules for dependency management
- Key dependencies include:
  - Kubernetes client-go for cluster interaction
  - Cobra for CLI framework
  - Various Kubernetes API libraries
  - JSON diff libraries for output formatting

## Validation Scenarios

### After Making Code Changes
1. **ALWAYS** run `make ensure` if dependencies changed
2. **ALWAYS** run `make build` -- ensure it completes without errors in ~1-2 minutes
3. **ALWAYS** run `make test_all` -- ensure tests pass in ~15 seconds
4. **ALWAYS** test `./kubespy --help` and `./kubespy version` to verify basic functionality
5. **ALWAYS** test all modified subcommands with --help flag

### Integration Testing
- kubespy requires a Kubernetes cluster to test full functionality
- Example YAML files are available in examples/ directories for testing
- Basic validation can be done without cluster by checking help output and version

## Important Notes
- This is a CLI tool, not a web application
- Primary entry point is kubespy.go which calls cmd.Execute()
- All CLI commands are implemented in cmd/ directory using Cobra framework
- The application watches Kubernetes resources in real-time when connected to a cluster
- Build artifacts include a single binary named "kubespy"
- NEVER CANCEL long-running builds or tests - they have appropriate timeouts built-in