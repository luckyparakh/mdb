package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // This field will not be shown to user
	Version   int32     `json:"version"`
	Title     string    `json:"title" binding:"required,min=1,max=255"`
	Year      int32     `json:"year" binding:"required,yearrange"`
	Runtime   Runtime   `json:"runtime" binding:"required"`
	Genres    []string  `json:"genres" binding:"required,unique"`
}

type ListMovie struct {
	Title  string   `form:"title" binding:"omitempty,min=2,max=255"`
	Genres []string `form:"genres" binding:"omitempty,genre"`
	Filters
}

type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(string, []string, Filters) ([]*Movie, *Metadata, error)
	}
}

func NewModel(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}

func NewMockModel(db *sql.DB) Models {
	return Models{
		Movies: MockMovieModel{},
	}
}
