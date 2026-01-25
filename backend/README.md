# Backend Development Guide

## Linting

This project uses [golangci-lint](https://golangci-lint.run/) for code linting. The configuration is stored in `.golangci.yml`.

### Running the linter

To run the linter on the backend code:

```bash
# From the project root
make lint-backend

# Or directly from the backend directory
cd backend && golangci-lint run
```

### Installing golangci-lint

If you don't have golangci-lint installed, you can install it with:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Alternatively, you can install it as part of the setup process:

```bash
make setup
```

### Configuration

The linter is configured with various checks in `.golangci.yml`. The configuration includes:

- Formatting checks (gofmt, goimports)
- Code quality checks (govet, errcheck, staticcheck)
- Style checks (golint, golint, stylecheck)
- Security checks (gosec)
- Performance checks (gosimple, ineffassign)
- And many more

If you need to adjust the linting rules, modify the `.golangci.yml` file.

## Other Commands

- Run all tests: `make test-backend`
- Run all linters: `make lint` (for all components)
- Run all tests: `make test` (for all components)
