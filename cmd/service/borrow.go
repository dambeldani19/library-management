package service

import (
	"context"
	"errors"
	"time"

	"go-grpc/helpers"
	"go-grpc/model"
	pb "go-grpc/pb/library"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BorrowingServiceServer struct {
	pb.UnimplementedBorrowingServiceServer
	DB *gorm.DB
}

// CreateBorrowingTransaction(context.Context, *CreateBorrowingTransactionRequest) (*BorrowingTransactionResponse, error)
// func (s *BorrowingServiceServer) CreateBorrowingTransaction(ctx context.Context, req *pb.CreateBorrowingTransactionRequest) (*pb.BorrowingTransactionResponse, error) {

// 	userID, _, err := helpers.GetData(ctx)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
// 	}

// 	// Parsing string ke time.Time
// 	layout := "2006-01-02 15:04:05"

// 	borrowingTransaction := model.BorrowingTransaction{
// 		BorrowerID: int32(userID),
// 		BookID:     req.BookId,
// 		BorrowedAt: time.Now().Format(layout),
// 		DueDate:    req.DueDate,
// 		Status:     "borrowed",
// 	}

// 	// Save the transaction to the database
// 	if err := s.DB.Create(&borrowingTransaction).Error; err != nil {
// 		return nil, err
// 	}

//		return &pb.BorrowingTransactionResponse{
//			Data: &pb.BorrowingTransaction{
//				Id:         borrowingTransaction.ID,
//				Borrower:   &pb.Borrower{Id: borrowingTransaction.BorrowerID},
//				Book:       &pb.Book{Id: borrowingTransaction.BookID},
//				BorrowedAt: borrowingTransaction.BorrowedAt,
//				DueDate:    borrowingTransaction.DueDate,
//				Status:     borrowingTransaction.Status,
//			},
//		}, nil
//	}
func (s *BorrowingServiceServer) CreateBorrowingTransaction(ctx context.Context, req *pb.CreateBorrowingTransactionRequest) (*pb.BorrowingTransactionResponse, error) {

	userID, _, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	// Parsing string ke time.Time
	layout := "2006-01-02 15:04:05"

	borrowingTransaction := model.BorrowingTransaction{
		BorrowerID: int32(userID),
		BookID:     req.BookId,
		BorrowedAt: time.Now().Format(layout),
		DueDate:    req.DueDate,
		Status:     "borrowed",
	}

	// Start a transaction
	tx := s.DB.Begin()

	// Save the transaction to the database
	if err := tx.Create(&borrowingTransaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update the book stock
	var bookStock model.BookStock
	if err := tx.Where("book_id = ?", req.BookId).First(&bookStock).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "book stock not found")
		}
		return nil, err
	}

	if bookStock.TotalStock <= 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.InvalidArgument, "insufficient book stock")
	}

	// Decrease the stock
	bookStock.TotalStock -= 1
	if err := tx.Save(&bookStock).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &pb.BorrowingTransactionResponse{
		Data: &pb.BorrowingTransaction{
			Id:         borrowingTransaction.ID,
			Borrower:   &pb.Borrower{Id: borrowingTransaction.BorrowerID},
			Book:       &pb.Book{Id: borrowingTransaction.BookID},
			BorrowedAt: borrowingTransaction.BorrowedAt,
			DueDate:    borrowingTransaction.DueDate,
			Status:     borrowingTransaction.Status,
		},
	}, nil
}

// GetBorrowingTransaction(context.Context, *IdRequest) (*BorrowingTransactionResponse, error)
func (s *BorrowingServiceServer) GetBorrowingTransaction(ctx context.Context, req *pb.IdRequest) (*pb.BorrowingTransactionResponse, error) {
	var borrowingTransaction model.BorrowingTransaction

	userID, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	// Query untuk admin atau user
	query := s.DB.Preload("Borrower").
		Preload("Book").
		Preload("Book.Author").
		Preload("Book.Category").
		Where("id = ?", req.Id)

	if role != "admin" {
		query = query.Where("borrower_id = ?", userID)
	}

	// Temukan transaksi sesuai query
	if err := query.First(&borrowingTransaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction not found or does not belong to user")
		}
		return nil, err
	}

	// Convert to protobuf message and return
	return &pb.BorrowingTransactionResponse{
		Data: &pb.BorrowingTransaction{
			Id: borrowingTransaction.ID,
			Borrower: &pb.Borrower{
				Id:    borrowingTransaction.BorrowerID,
				Name:  borrowingTransaction.Borrower.Name,
				Email: borrowingTransaction.Borrower.Name,
			},
			Book: &pb.Book{
				Id:              borrowingTransaction.BookID,
				Title:           borrowingTransaction.Book.Title,
				PublicationYear: borrowingTransaction.Book.PublicationYear,
				Description:     borrowingTransaction.Book.Description,
				Author: &pb.Author{
					Id:   borrowingTransaction.Book.AuthorID,
					Name: borrowingTransaction.Book.Author.Name,
					Bio:  borrowingTransaction.Book.Author.Bio,
				},
				Category: &pb.Category{
					Id:          borrowingTransaction.Book.CategoryID,
					Name:        borrowingTransaction.Book.Category.Name,
					Description: borrowingTransaction.Book.Category.Description,
				},
			},
			BorrowedAt: borrowingTransaction.BorrowedAt,
			DueDate:    borrowingTransaction.DueDate,
			ReturnedAt: borrowingTransaction.ReturnedAt.String,
			Status:     borrowingTransaction.Status,
		},
	}, nil
}

// UpdateBorrowingTransaction(context.Context, *UpdateBorrowingTransactionRequest) (*BorrowingTransactionResponse, error)
func (s *BorrowingServiceServer) UpdateBorrowingTransaction(ctx context.Context, req *pb.UpdateBorrowingTransactionRequest) (*pb.BorrowingTransactionResponse, error) {
	var existingTransaction model.BorrowingTransaction

	userID, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	// Query untuk admin atau user
	query := s.DB.Where("id = ?", req.Id)
	if role != "admin" {
		query = query.Where("borrower_id = ?", userID)
	}

	// Temukan transaksi yang sesuai
	if err := query.First(&existingTransaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "transaction not found or does not belong to user")
		}
		return nil, err
	}

	existingTransaction.BookID = req.BookId
	existingTransaction.BorrowerID = req.BorrowerId
	existingTransaction.DueDate = req.DueDate
	existingTransaction.ReturnedAt.String = req.ReturnedAt
	existingTransaction.Status = req.Status

	// Save the changes
	if err := s.DB.Save(&existingTransaction).Error; err != nil {
		return nil, err
	}

	return &pb.BorrowingTransactionResponse{
		Data: &pb.BorrowingTransaction{
			Id:         existingTransaction.ID,
			Borrower:   &pb.Borrower{Id: existingTransaction.BorrowerID},
			Book:       &pb.Book{Id: existingTransaction.BookID},
			BorrowedAt: existingTransaction.BorrowedAt,
			DueDate:    existingTransaction.DueDate,
			ReturnedAt: existingTransaction.ReturnedAt.String,
			Status:     existingTransaction.Status,
		},
	}, nil
}

// ListBorrowingTransactions(context.Context, *Empty) (*BorrowingTransactionsResponse, error)
func (s *BorrowingServiceServer) ListBorrowingTransactions(ctx context.Context, req *pb.Empty) (*pb.BorrowingTransactionsResponse, error) {
	var transactions []model.BorrowingTransaction

	userID, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	// Query untuk admin atau user
	query := s.DB.Preload("Borrower").
		Preload("Book").
		Preload("Book.Author").
		Preload("Book.Category").
		Where("1 = 1")

	if role != "admin" {
		query = query.Where("borrowing_transactions.borrower_id = ?", userID)
	}

	// Ambil semua transaksi sesuai query
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}

	var pbTransactions []*pb.BorrowingTransaction
	for _, transaction := range transactions {
		pbTransaction := &pb.BorrowingTransaction{
			Id: transaction.ID,
			Borrower: &pb.Borrower{
				Id:    transaction.BorrowerID,
				Name:  transaction.Borrower.Name,
				Email: transaction.Borrower.Name,
			},
			Book: &pb.Book{
				Id:              transaction.BookID,
				Title:           transaction.Book.Title,
				PublicationYear: transaction.Book.PublicationYear,
				Description:     transaction.Book.Description,
				Author: &pb.Author{
					Id:   transaction.Book.AuthorID,
					Name: transaction.Book.Author.Name,
					Bio:  transaction.Book.Author.Bio,
				},
				Category: &pb.Category{
					Id:          transaction.Book.CategoryID,
					Name:        transaction.Book.Category.Name,
					Description: transaction.Book.Category.Description,
				},
			},
			BorrowedAt: transaction.BorrowedAt,
			DueDate:    transaction.DueDate,
			ReturnedAt: transaction.ReturnedAt.String,
			Status:     transaction.Status,
		}
		pbTransactions = append(pbTransactions, pbTransaction)
	}

	return &pb.BorrowingTransactionsResponse{
		Data: pbTransactions,
	}, nil
}
