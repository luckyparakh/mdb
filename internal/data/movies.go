package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

const timeout = 3

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version` // Returning is PSQL syntax
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, fmt.Errorf("record not found")
	}

	// To mimic timeout at DB
	// Use curl -w '\nTime: %{time_total}s \n' localhost:4000/v1/movies/2
	// qry := `SELECT pg_sleep(5), id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE id = $1`

	qry := `SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`
	var movie Movie
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, qry, id).Scan(
		// &[]byte{},
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("record not found")
		default:
			return nil, err
		}
	}
	return &movie, nil
}
func (m MovieModel) Update(movie *Movie) error {
	// https://stackoverflow.com/questions/129329/optimistic-vs-pessimistic-locking/129397#129397
	// By adding version in where clause, is way to stop read condition via optimistic locking
	query := `UPDATE movies
				SET title = $1, year = $2, runtime = $3, genres = $4, version =
				version + 1
				WHERE id = $5 and version = $6
				RETURNING version`
	args := []interface{}{
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.ID,
		&movie.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
}
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return fmt.Errorf("record not found")
	}
	query := `DELETE FROM movies WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	r, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, *Metadata, error) {
	movies := []*Movie{}
	tr := 0
	qry := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY %v, id ASC
	LIMIT $3 OFFSET $4`, filters.sortCol())
	// qry := `SELECT id, created_at, title, year, runtime, genres, version FROM movies ORDER BY id`
	// log.Println(qry)
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}
	log.Println(args)
	rows, err := m.DB.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&tr,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, nil, err
		}
		movies = append(movies, &movie)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	metadata := calculateMetadata(tr, filters.Page, filters.PageSize)
	return movies, &metadata, nil
}

type MockMovieModel struct{}

func (m MockMovieModel) Insert(movie *Movie) error {
	return nil
}
func (m MockMovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}
func (m MockMovieModel) Update(movie *Movie) error {
	return nil
}
func (m MockMovieModel) Delete(id int64) error {
	return nil
}
func (m MockMovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, *Metadata, error) {
	return nil, nil, nil
}
