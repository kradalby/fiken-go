// Package ops is the domain operation layer over the ogen-generated
// Fiken client. Each file groups operations by Fiken API tag.
//
// `ops/mutating.gen.go` is generated from api/fiken-openapi.yaml by
// cmd/fiken-mutating-gen. DO NOT hand-edit.
package ops

//go:generate go run ../cmd/fiken-mutating-gen -spec ../api/fiken-openapi.yaml -out mutating.gen.go
