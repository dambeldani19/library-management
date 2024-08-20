package main

import (
	"log"
	"net"

	"go-grpc/cmd/config"
	"go-grpc/cmd/service"
	"go-grpc/middleware"
	libraryPb "go-grpc/pb/library"

	"google.golang.org/grpc"
)

func main() {

	port := ":50051"
	netListen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen %v", err.Error())
	}

	db := config.ConnectDatabase()

	// Create gRPC server with JWT middleware interceptor
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.JWTMiddleware(db)))

	// Register services
	authService := service.AuthService{DB: db}
	libraryPb.RegisterAuthServiceServer(grpcServer, &authService)

	bookService := service.BookService{DB: db}
	libraryPb.RegisterBookServiceServer(grpcServer, &bookService)

	categoryService := service.CategoryService{DB: db}
	libraryPb.RegisterCategoryServiceServer(grpcServer, &categoryService)

	authorService := service.AuthorService{DB: db}
	libraryPb.RegisterAuthorServiceServer(grpcServer, &authorService)

	stockService := service.BookStockService{DB: db}
	libraryPb.RegisterBookStockServiceServer(grpcServer, &stockService)

	borrowService := service.BorrowingServiceServer{DB: db}
	libraryPb.RegisterBorrowingServiceServer(grpcServer, &borrowService)

	returnedService := service.ReturningServiceServer{DB: db}
	libraryPb.RegisterReturningServiceServer(grpcServer, &returnedService)

	log.Printf("Server start at %v", netListen.Addr())
	if err := grpcServer.Serve(netListen); err != nil {
		log.Fatalf("failed to serve %v", err.Error())
	}
}
