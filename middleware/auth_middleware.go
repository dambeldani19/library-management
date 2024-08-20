package middleware

import (
	"context"
	"go-grpc/helpers"
	"strings"

	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var jwtSecretKey = []byte("your-secret-key")

func JWTMiddleware(db *gorm.DB) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// Bypass JWT middleware for AuthService methods
		if strings.Contains(info.FullMethod, "AuthService") {
			return handler(ctx, req)
		}

		// Implement JWT verification for other services
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(http.StatusUnauthorized, "missing metadata")
		}

		authHeader, exists := md["authorization"]
		if !exists || len(authHeader) == 0 {
			return nil, status.Errorf(http.StatusUnauthorized, "authorization token is not supplied")
		}

		tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
		userId, role, err := helpers.ValidateToken(tokenStr)
		if err != nil {
			return nil, status.Errorf(http.StatusUnauthorized, "invalid token")
		}
		us := *userId
		r := *role
		// Set values into context
		ctx = context.WithValue(ctx, "userID", us)
		ctx = context.WithValue(ctx, "role", r)

		return handler(ctx, req)
	}
}
