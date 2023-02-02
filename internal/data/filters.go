package data

import (
	"fmt"
	"math"
	"strings"
)

type Filters struct {
	Page     int    `form:"page" binding:"omitempty,min=1,max=10000"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Sort     string `form:"sort" binding:"omitempty,oneof=id title year runtime -id -title -year -runtime"`
}

// Define a new Metadata struct for holding the pagination metadata.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func (f *Filters) sortCol() string {
	if strings.HasPrefix(f.Sort, "-") {
		return fmt.Sprintf("%v DESC", strings.TrimPrefix(f.Sort, "-"))
	}
	return fmt.Sprintf("%v ASC", f.Sort)
}

func (f *Filters) limit() int {
	// log.Println(f.PageSize)
	return f.PageSize
}

func (f *Filters) offset() int {
	// log.Println((f.Page - 1) * f.PageSize)
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(total, page, page_size int) Metadata {
	if total == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     page_size,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(total) / float64(page_size))),
		TotalRecords: total,
	}
}
