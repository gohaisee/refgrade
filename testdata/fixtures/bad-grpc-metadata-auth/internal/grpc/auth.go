package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func UserFromContext(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	vals := md.Get("x-user-id")
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}
