package data

import (
	"database/sql"
	"time"
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
type IUser interface {
	Update(*User) error
	GetByEmail(string) (*User, error)
	Insert(*User) error
	GetForToken(tokenScope, tokenPlaintext string) (*User, error)
}

type IToken interface {
	Delete(int64, string) error
	Insert(*Token) error
	New(int64, time.Duration, string) (*Token, error)
}
type IPermission interface {
	GetAllForUser(userID int64) (Permissions, error)
	AddForUser(userID int64, codes ...string) error
}
type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(string, []string, Filters) ([]*Movie, *Metadata, error)
	}
	User       IUser
	Token      IToken
	Permission IPermission
}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name" binding:"required,min=2,max=255"`
	Email     string    `json:"email" binding:"required,email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}
type password struct {
	plaintext *string `json:"-" binding:"required,min=2,max=255"`
	hash      []byte  `json:"-" binding:"required"`
}

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expire"`
	Scope     string    `json:"-"`
}

func NewModel(db *sql.DB) Models {
	return Models{
		Movies:     MovieModel{DB: db},
		User:       UserModel{DB: db},
		Token:      TokenModel{DB: db},
		Permission: PermissionModel{DB: db},
	}
}

func NewMockModel(db *sql.DB) Models {
	return Models{
		Movies: MockMovieModel{},
	}
}
