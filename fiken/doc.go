// Package fiken is the ogen-generated typed client and server skeleton
// for the Fiken REST API. All exported types and methods in this
// package are produced by `go generate ./fiken` from
// `../api/fiken-openapi.yaml` and the config in `ogen.yml`.
//
// DO NOT hand-edit any file in this package. The pre-commit hook
// `codegen-clean` re-runs `go generate ./...` and fails the commit if
// the working tree differs from the regenerated output.
package fiken

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target . --package fiken --clean ../api/fiken-openapi.yaml
