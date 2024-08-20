package helpers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetData(ctx context.Context) (int, string, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return 0, "", status.Errorf(codes.InvalidArgument, "userID not found")
	}

	role, ok := ctx.Value("role").(string)
	if !ok {
		return 0, "", status.Errorf(codes.InvalidArgument, "role not found")
	}

	return userID, role, nil

}
