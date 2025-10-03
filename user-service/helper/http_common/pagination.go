package httpcommon

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Paging struct {
	Offset     int   `json:"offset"`
	Limit      int   `json:"limit"`
	TotalCount int64 `json:"totalCount"`
}

type Meta struct {
	TotalCount     int64 `json:"totalCount"`
	IsLastPage     bool  `json:"isLastPage"`
	NumPage        int   `json:"numPage"`
	AdditionalData any   `json:"additionalData,omitempty"`
}

const pageSize = "pageSize"

const currentPage = "pageIndex"

func ParseParams(c *gin.Context) *Paging {
	var p Paging

	if pageSize, err := strconv.Atoi(c.Query(pageSize)); err == nil && pageSize > 0 {
		p.Limit = pageSize
	}

	if p.Limit == 0 {
		p.Limit = 10
	}

	if currentPage, err := strconv.Atoi(c.Query(currentPage)); err == nil && currentPage > 0 {
		p.Offset = (currentPage - 1) * p.Limit
	}

	return &p
}

type Filters map[string]string

func ParseFilters(c *gin.Context) Filters {
	filters := Filters{}
	for key, value := range c.Request.URL.Query() {
		if key != pageSize && key != currentPage {
			filters[key] = value[0]
		}
	}
	return filters
}

type SortDirection string

const (
	Ascend  SortDirection = "ascend"
	Descend SortDirection = "descend"
)

const (
	SortDirectionAscend  = "asc"
	SortDirectionDescend = "desc"
)

func GetCurrentPage(c *gin.Context) int {
	pageIndex, err := strconv.Atoi(c.Query(currentPage))

	if pageIndex == 0 || err != nil {
		return 1
	}

	return pageIndex
}
