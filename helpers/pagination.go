package helpers

import (
	"go-grpc/pb/pagination"
	"math"

	"gorm.io/gorm"
)

func Pagination(sql *gorm.DB, page, limit int64, pagination *pagination.Pagination) (int64, int64) {
	var total int64
	var offset int64

	sql.Count(&total)

	if page == 1 {
		offset = 0
	} else {
		offset = (page - 1) * limit
	}

	pagination.Total = uint64(total)
	pagination.PerPage = uint32(limit)
	pagination.CurrentPage = uint32(page)
	pagination.Lastpage = uint32(math.Ceil(float64(total) / float64(limit)))

	return offset, limit
}
