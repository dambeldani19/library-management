package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-grpc/model" // Import the models package
	pb "go-grpc/pb/library"

	"gorm.io/gorm"
)

type ReturningServiceServer struct {
	pb.UnimplementedReturningServiceServer
	DB *gorm.DB
}

func (s *ReturningServiceServer) ReturnBook(ctx context.Context, req *pb.ReturnBookRequest) (*pb.ReturnBookResponse, error) {
	var transaction model.BorrowingTransaction

	// Temukan transaksi berdasarkan ID
	err := s.DB.First(&transaction, req.TransactionId).Error
	if err != nil {
		if gorm.ErrRecordNotFound == err {
			return &pb.ReturnBookResponse{Success: false, Message: "Transaction not found"}, nil
		}
		return nil, err
	}

	if transaction.ReturnedAt.String != "" {
		return &pb.ReturnBookResponse{Success: false, Message: "Book already returned"}, nil
	}

	// Parse timestamp
	layout := "2006-01-02 15:04:05"
	returnAt, err := time.Parse(layout, req.ReturnedAt)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return nil, err
	}

	transaction.ReturnedAt = sql.NullString{String: req.ReturnedAt, Valid: true}
	transaction.Status = "returned"

	// Simpan perubahan
	if err := s.DB.Save(&transaction).Error; err != nil {
		return nil, err
	}

	// Simpan informasi pengembalian di tabel returning_transactions
	returningTransaction := model.ReturningTransaction{
		BorrowingTransactionID: transaction.ID,
		ReturnedAt:             returnAt,
	}

	if err := s.DB.Create(&returningTransaction).Error; err != nil {
		return nil, err
	}

	return &pb.ReturnBookResponse{Success: true, Message: "Book returned successfully"}, nil
}
