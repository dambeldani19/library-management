package service

import (
	"fmt"
	"go-grpc/helpers"
	pb "go-grpc/pb/library"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BookStockService struct {
	pb.UnimplementedBookStockServiceServer
	DB *gorm.DB
}

// GetBookStock(context.Context, *BookStockRequest) (*BookStockResponse, error)
func (s *BookStockService) GetBookStock(ctx context.Context, req *pb.IdRequest) (*pb.BookStockResponse, error) {

	row := s.DB.Table("books as b").
		Joins("LEFT JOIN authors au on au.id = b.author_id").
		Joins("LEFT JOIN categories c on c.id = b.category_id").
		Joins("LEFT JOIN book_stocks bs on bs.book_id = b.id").
		Select("b.id, b.title,b.publication_year, b.description, au.id, au.name, au.bio, c.id, c.name category_name, c.description, bs.id, bs.total_stock").
		Where("b.id = ?", req.GetId()).
		Row()

	var book pb.Book
	var author pb.Author
	var category pb.Category
	var stock pb.BookStock

	if err := row.Scan(&book.Id, &book.Title, &book.PublicationYear, &book.Description, &author.Id,
		&author.Name, &author.Bio, &category.Id, &category.Name, &category.Description, &stock.Id, &stock.TotalStock); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	book.Author = &author
	book.Category = &category
	stock.Book = &book

	stockRes := &pb.BookStockResponse{
		Data: &stock,
	}

	return stockRes, nil

}

// UpdateBookStock(context.Context, *BookStockUpdate) (*BookStockResponse, error)
func (s *BookStockService) UpdateBookStock(ctx context.Context, req *pb.BookStockUpdate) (*pb.BookStockResponse, error) {

	_, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	if role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "no access for borrower: %v", err)
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Table("book_stocks").
			Where("book_id = ?", req.BookId).
			Update("total_stock", req.TotalStock)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("no ParamRequests affected")
		}

		return nil

	})

	return nil, err
}
