package service

import (
	"go-grpc/helpers"
	pb "go-grpc/pb/library"
	paginationPb "go-grpc/pb/pagination"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BookService struct {
	pb.UnimplementedBookServiceServer
	DB *gorm.DB
}

// ListBooks(context.Context, *ParameterReq) (*BooksResponse, error)
func (s *BookService) ListBooks(ctx context.Context, req *pb.ParameterReq) (*pb.BooksResponse, error) {

	var books []*pb.Book
	var pagination paginationPb.Pagination

	sql := s.DB.Table("books as b").
		Joins("LEFT JOIN authors au on au.id = b.author_id").
		Joins("LEFT JOIN categories c on c.id = b.category_id").
		Select("b.id, b.title,b.publication_year, b.description, au.id, au.name, au.bio, c.id, c.name category_name, c.description")

	offset, limit := helpers.Pagination(sql, req.Page, req.Limit, &pagination)

	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var book pb.Book
		var author pb.Author
		var category pb.Category

		if err := rows.Scan(&book.Id, &book.Title, &book.PublicationYear, &book.Description, &author.Id, &author.Name, &author.Bio, &category.Id, &category.Name, &category.Description); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		book.Author = &author
		book.Category = &category

		books = append(books, &book)
	}

	booksRes := &pb.BooksResponse{
		Pagination: &pagination,
		Data:       books,
	}

	return booksRes, nil
}

// GetBook(context.Context, *BookRequest) (*BookResponse, error)
func (s *BookService) GetBook(ctx context.Context, req *pb.BookRequest) (*pb.BookResponse, error) {

	row := s.DB.Table("books as b").
		Joins("LEFT JOIN authors au on au.id = b.author_id").
		Joins("LEFT JOIN categories c on c.id = b.category_id").
		Select("b.id, b.title,b.publication_year, b.description, au.id, au.name, au.bio, c.id, c.name category_name, c.description").
		Where("b.id = ?", req.GetId()).
		Row()

	var book pb.Book
	var author pb.Author
	var category pb.Category

	if err := row.Scan(&book.Id, &book.Title, &book.PublicationYear, &book.Description, &author.Id, &author.Name, &author.Bio, &category.Id, &category.Name, &category.Description); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	book.Author = &author
	book.Category = &category

	bookRes := &pb.BookResponse{
		Data: &book,
	}

	return bookRes, nil

}

// CreateBook(context.Context, *Book) (*BookResponse, error)
func (s *BookService) CreateBook(ctx context.Context, book *pb.CreateBookRequest) (*pb.BookResponse, error) {

	_, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	if role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "no access for borrower: %v", err)
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		category := pb.Category{
			Id:          0,
			Name:        book.GetCategory().GetName(),
			Description: book.GetCategory().GetDescription(),
		}

		if err := tx.Table("categories").Where("LCASE(name) = ?", category.GetName()).FirstOrCreate(&category).Error; err != nil {
			return err
		}

		b := struct {
			Title           string
			Description     string
			AuthorID        uint64
			CategoryID      uint64
			PublicationYear uint32
		}{
			Title:           book.GetTitle(),
			Description:     book.GetDescription(),
			AuthorID:        uint64(book.GetAuthorId()),
			CategoryID:      uint64(category.GetId()),
			PublicationYear: uint32(book.PublicationYear),
		}

		if err := tx.Table("books").Create(&b).Error; err != nil {
			return err
		}

		return nil

	})

	return nil, err
}

// UpdateBook(context.Context, *BookUpdateReq) (*BookResponse, error)
func (s *BookService) UpdateBook(ctx context.Context, req *pb.BookUpdateReq) (*pb.BookResponse, error) {

	_, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	if role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "no access for borrower: %v", err)
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {

		b := struct {
			Title           string
			Description     string
			AuthorID        uint64
			CategoryID      uint64
			PublicationYear uint32
		}{
			Title:           req.GetTitle(),
			Description:     req.GetDescription(),
			PublicationYear: uint32(req.PublicationYear),
		}

		if err := tx.Table("books").Where("id = ?", req.Id).Updates(&b).Error; err != nil {
			return err
		}

		return nil

	})

	return nil, err
}

// DeleteBook(context.Context, *BookRequest) (*Empty, error)
func (s *BookService) DeleteBook(ctx context.Context, req *pb.BookRequest) (*pb.Empty, error) {

	_, role, err := helpers.GetData(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user data: %v", err)
	}

	if role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "no access for borrower: %v", err)
	}

	if err = s.DB.Table("books").Where("id = ?", req.Id).Delete(nil).Error; err != nil {
		return nil, err
	}

	return nil, nil
}
