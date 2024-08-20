package service

import (
	"go-grpc/errorhandler"
	"go-grpc/helpers"
	pb "go-grpc/pb/library"
	"time"

	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	DB *gorm.DB
}

// RegisterBorrower(context.Context, *RegisterUser) (*ReturnSimpleResponse, error)
func (s *AuthService) RegisterBorrower(ctx context.Context, req *pb.RegisterUser) (*pb.ReturnSimpleResponse, error) {

	passwordHash, err := helpers.HashPassword(req.Password)
	if err != nil {
		return &pb.ReturnSimpleResponse{
			Success: false,
			Message: "Failed to register borrower",
		}, err
	}

	borrower := struct {
		Name      string
		Email     string
		Password  string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		Name:      req.Name,
		Email:     req.Email,
		Password:  passwordHash, // Hash password in real implementation
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.DB.Table("borrowers").Create(&borrower).Error; err != nil {
		return &pb.ReturnSimpleResponse{
			Success: false,
			Message: "Failed to register borrower",
		}, err
	}

	return &pb.ReturnSimpleResponse{
		Success: true,
		Message: "Borrower registered successfully",
	}, nil
}

// RegisterAdmin(context.Context, *RegisterUser) (*ReturnSimpleResponse, error)
func (s *AuthService) RegisterAdmin(ctx context.Context, req *pb.RegisterUser) (*pb.ReturnSimpleResponse, error) {
	passwordHash, err := helpers.HashPassword(req.Password)
	if err != nil {
		return &pb.ReturnSimpleResponse{
			Success: false,
			Message: "Failed to register borrower",
		}, err
	}

	admin := struct {
		Name      string
		Email     string
		Password  string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		Name:      req.Name,
		Email:     req.Email,
		Password:  passwordHash, // Hash password in real implementation
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.DB.Table("admin").Create(&admin).Error; err != nil {
		return &pb.ReturnSimpleResponse{
			Success: false,
			Message: "Failed to register admin",
		}, err
	}

	return &pb.ReturnSimpleResponse{
		Success: true,
		Message: "Admin registered successfully",
	}, nil
}

// Login method implementation
func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.ResponseParamLogin, error) {
	var adminOrBorrower struct {
		ID       int32
		Name     string
		Email    string
		Password string
	}
	role := "admin"

	// Check for user in Admin table
	if err := s.DB.Table("admin").Where("email = ?", req.Email).First(&adminOrBorrower).Error; err != nil {
		// If not found, check in Borrowers table
		if err := s.DB.Table("borrowers").Where("email = ?", req.Email).First(&adminOrBorrower).Error; err != nil {
			return &pb.ResponseParamLogin{
				StatusCode: 404,
				Message:    "User not found",
				Data:       nil,
			}, nil
		}
		if adminOrBorrower.Email != "" {
			role = "borrower"

		}
	}

	if err := helpers.VerifyPassword(adminOrBorrower.Password, req.Password); err != nil {
		return nil, &errorhandler.NotFoundError{Message: "wrong email or password"}
	}

	token, err := helpers.GenerateToken(int(adminOrBorrower.ID), role)
	if err != nil {
		return nil, &errorhandler.InternalServerError{Message: err.Error()}
	}

	return &pb.ResponseParamLogin{
		StatusCode: 200,
		Message:    "Login successful",
		Data: &pb.LoginResponse{
			Id:    adminOrBorrower.ID,
			Name:  adminOrBorrower.Name,
			Token: token,
		},
	}, nil
}
