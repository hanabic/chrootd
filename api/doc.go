//go:generate protoc --go_out=plugins=grpc:. containerpool.proto container.proto image.proto
package api
