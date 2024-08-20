package service

import (
	"database/sql"
	"fmt"
	"go-grpc/helpers"
	pb "go-grpc/pb/library"
	paginationPb "go-grpc/pb/pagination"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthorService struct {
	pb.UnimplementedAuthorServiceServer
	DB *gorm.DB
}

// ListAuthors(context.Context, *ParameterReq) (*AuthorsResponse, error)
func (s *AuthorService) ListAuthors(ctx context.Context, req *pb.ParameterReq) (*pb.AuthorsResponse, error) {

	var authors []*pb.Author
	var pagination paginationPb.Pagination

	sql := s.DB.Table("authors as a").
		Select("a.id, a.name, a.bio")

	offset, limit := helpers.Pagination(sql, req.Page, req.Limit, &pagination)

	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var author pb.Author

		if err := rows.Scan(&author.Id, &author.Name, &author.Bio); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		authors = append(authors, &author)
	}

	categoryRes := &pb.AuthorsResponse{
		Pagination: &pagination,
		Data:       authors,
	}

	return categoryRes, nil
}

// GetAuthor(context.Context, *IdRequest) (*AuthorResponse, error)
func (s *AuthorService) GetAuthor(ctx context.Context, req *pb.IdRequest) (*pb.AuthorResponse, error) {

	row := s.DB.Table("authors as a").
		Select("a.id, a.name, a.bio").
		Where("a.id = ?", req.GetId()).
		Row()

	var auhtor pb.Author

	if err := row.Scan(&auhtor.Id, &auhtor.Name, &auhtor.Bio); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	auhtorRes := &pb.AuthorResponse{
		Data: &auhtor,
	}

	return auhtorRes, nil

}

// UpdateAuthor(context.Context, *Author) (*AuthorResponse, error)
func (s *AuthorService) UpdateAuthor(ctx context.Context, req *pb.Author) (*pb.AuthorResponse, error) {

	author := pb.Author{
		Id:   req.GetId(),
		Name: req.GetName(),
		Bio:  req.GetBio(),
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Table("authors").Where("id = ?", author.GetId()).Updates(&author)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("no ParamRequests affected")
		}

		return nil

	})

	if err != nil {
		return nil, err
	}

	authorRes := &pb.AuthorResponse{
		Data: &author,
	}

	return authorRes, err
}

// DeleteAuthor(context.Context, *IdRequest) (*Empty, error)
func (s *AuthorService) DeleteAuthor(ctx context.Context, req *pb.IdRequest) (*pb.Empty, error) {

	q := s.DB.Table("books as b").
		Select("b.id").
		Where("author_id = ?", req.GetId()).
		Row()

	var book pb.Book

	if err := q.Scan(&book.Id); err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	if book.Id > 0 {
		return nil, status.Error(codes.Canceled, "author id sudah tercantum di book, tidak bisa di hapus")
	} else {
		if err := s.DB.Table("authors").Where("id = ?", req.Id).Delete(nil).Error; err != nil {
			return nil, err
		}
	}
	return nil, nil
}
