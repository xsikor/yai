# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `go build`
- Run tests: `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Lint: Uses golangci-lint with thelper, gofumpt, tparallel, unconvert linters

## Code Style
- Packages: Small focused packages (ai, config, system, ui, etc.)
- Imports: Group standard library, 3rd party, and internal imports
- Error handling: Return errors up the call stack, use fmt.Errorf for context
- Naming: CamelCase for exported symbols, camelCase for unexported
- Method chaining pattern with pointer receivers returning `*Type`
- Use interfaces for abstractions and testing
- Struct field comments not required, prefer self-documenting names

## Project Structure
- Each package has corresponding test file with same name
- Test function naming: `TestPackageName` or `TestTypeName`
- Prefer Go's standard error handling over custom error types