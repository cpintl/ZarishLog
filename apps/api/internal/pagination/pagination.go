package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type OffsetLimit struct {
	Offset int
	Limit  int
}

const DefaultPage = 1
const DefaultPageSize = 20
const MaxPageSize = 200

func FromQuery(c *gin.Context) Params {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = DefaultPage
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Params{Page: page, PageSize: pageSize}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Params) Limit() int {
	return p.PageSize
}

func (p Params) OffsetLimit() OffsetLimit {
	return OffsetLimit{Offset: p.Offset(), Limit: p.Limit()}
}

func (p Params) TotalPages(total int) int {
	if p.PageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(p.PageSize)))
}
