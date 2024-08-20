package service

import (
	"context"
	"go-grpc/helpers"
	pb "go-grpc/pb/library"
	paginationPb "go-grpc/pb/pagination"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CategoryService struct {
	pb.UnimplementedCategoryServiceServer
	DB *gorm.DB
}

// ListCategories(context.Context, *ParameterReq) (*CategoriesResponse, error)
func (s *CategoryService) ListCategories(ctx context.Context, req *pb.ParameterReq) (*pb.CategoriesResponse, error) {

	var categories []*pb.Category
	var pagination paginationPb.Pagination

	sql := s.DB.Table("categories as c").
		Select("c.id, c.name category_name, c.description")

	offset, limit := helpers.Pagination(sql, req.Page, req.Limit, &pagination)

	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var category pb.Category

		if err := rows.Scan(&category.Id, &category.Name, &category.Description); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		categories = append(categories, &category)
	}

	categoryRes := &pb.CategoriesResponse{
		Pagination: &pagination,
		Data:       categories,
	}

	return categoryRes, nil
}

// GetCategory(context.Context, *CategoryRequest) (*CategoryResponse, error)
func (s *CategoryService) GetCategory(ctx context.Context, req *pb.IdRequest) (*pb.CategoryResponse, error) {

	row := s.DB.Table("categories as c").
		Select("c.id, c.name category_name, c.description").
		Where("c.id = ?", req.GetId()).
		Row()

	var category pb.Category

	if err := row.Scan(&category.Id, &category.Name, &category.Description); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	categoryRes := &pb.CategoryResponse{
		Data: &category,
	}

	return categoryRes, nil

}

// CreateCategory(context.Context, *Category) (*CategoryResponse, error)
func (s *CategoryService) CreateCategory(ctx context.Context, req *pb.CategoryRequest) (*pb.CategoryResponse, error) {

	category := pb.Category{
		Id:          0,
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Table("categories").Where("LCASE(name) = ?", category.GetName()).FirstOrCreate(&category).Error; err != nil {
			return err
		}

		return nil

	})

	categoryRes := &pb.CategoryResponse{
		Data: &category,
	}

	return categoryRes, err
}

// UpdateCategory(context.Context, *CategoryRequest) (*CategoryResponse, error)
func (s *CategoryService) UpdateCategory(ctx context.Context, req *pb.CategoryRequest) (*pb.CategoryResponse, error) {

	category := pb.Category{
		Id:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Table("categories").Where("id = ?", category.GetId()).Updates(&category).Error; err != nil {
			return err
		}

		return nil

	})

	categoryRes := &pb.CategoryResponse{
		Data: &category,
	}

	return categoryRes, err
}

// DeleteCategory(context.Context, *IdRequest) (*Empty, error)
func (s *CategoryService) DeleteCategory(ctx context.Context, req *pb.IdRequest) (*pb.Empty, error) {

	q := s.DB.Table("books as b").
		Select("b.id").
		Where("category_id = ?", req.GetId()).
		Row()

	var book pb.Book

	if err := q.Scan(&book.Id); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if book.Id > 0 {
		return nil, status.Error(codes.Canceled, "category id sudah tercantum di book, tidak bisa di hapus")
	} else {
		if err := s.DB.Table("categories").Where("id = ?", req.Id).Delete(nil).Error; err != nil {
			return nil, err
		}
	}

	return nil, nil
}
