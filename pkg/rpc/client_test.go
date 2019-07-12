package rpc

import (
	"context"
	"testing"
)

func TestPutNodeStorage(t *testing.T) {
	Init(":50051")
	err := PutNodeStorage(context.Background())
	t.Error(err)
}
