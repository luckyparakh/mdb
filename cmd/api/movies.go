package main

import (
	"database/sql"
	"fmt"
	"mdb/internal/data"
	"mdb/internal/validation"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (a *application) createMovieHandler(c *gin.Context) {
	var movie data.Movie
	// To do add validation
	// Like size of body, multiple json value input
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, envelope{"errors": validation.Errors(err)})
		return
	}
	if err := a.models.Movies.Insert(&movie); err != nil {
		c.JSON(http.StatusInternalServerError, a.createError(err, "Error while inserting movie"))
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	c.JSON(http.StatusOK, envelope{"movie": movie})
}

func (a *application) showMovieHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, a.createError(err, "Error while converting to int or id is less than 1"))
		return
	}
	movie, err := a.models.Movies.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &envelope{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, &envelope{"movie": movie})
	// c.IndentedJSON(http.StatusOK, &movie) // Will make output prety if used with curl command, but it will expensive than non indented one
}

func (a *application) updateMovieHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, a.createError(err, "Id should be a valid integer"))
		return
	}
	// Reason of creating one more struct is that feild type and validation are different than
	// that of movie struct.
	var input struct {
		Title   *string       `json:"title" binding:"omitempty,min=1,max=255"`
		Year    *int32        `json:"year" binding:"omitempty,yearrange"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres" binding:"unique"`
	}

	// To do add validation
	// Like size of body, multiple json value input
	if err := c.ShouldBindJSON(&input); err != nil {
		a.logger.PrintError(err, nil)
		c.JSON(http.StatusBadRequest, envelope{"errors": validation.Errors(err)})
		return
	}
	dbMovie, err := a.models.Movies.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &envelope{"error": err.Error()})
		return
	}
	// If the input.Title value is nil then we know that no corresponding "title" key
	// value pair was provided in the JSON request body. So we move on and leave the
	// movie record unchanged. Otherwise, we update the movie record with the new title
	// value. Importantly, because input.Title is a now a pointer to a string, we need
	// to dereference the pointer using the * operator to get the underlying value
	// before assigning it to our movie record.
	if input.Title != nil {
		dbMovie.Title = *input.Title
	}

	if input.Genres != nil {
		dbMovie.Genres = input.Genres
	}
	if input.Runtime != nil {
		dbMovie.Runtime = *input.Runtime
	}

	if input.Year != nil {
		dbMovie.Year = *input.Year
	}

	if err := a.models.Movies.Update(dbMovie); err != nil {
		// https://stackoverflow.com/questions/129329/optimistic-vs-pessimistic-locking/129397#129397
		// Below is way to stop read condition via optimistic locking
		if err == sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, a.createError(err, "unable to update the record due to an edit conflict, try again"))
			return
		}
		c.JSON(http.StatusInternalServerError, &envelope{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, &envelope{"movie": dbMovie})
}

func (a *application) deleteMovieHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, a.createError(err, "Id should be a valid integer"))
		return
	}
	err = a.models.Movies.Delete(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, a.createError(err, ""))
		return
	}
	msg := fmt.Sprintf("Record with ID:%d deleted.", id)
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func (a *application) listMoviesHandler(c *gin.Context) {
	var input data.ListMovie
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, validation.Errors(err))
		return
	}
	addDefaultValue(&input)
	// log.Println(input)
	mvs, md, err := a.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, a.createError(err, ""))
		return
	}
	c.JSON(http.StatusOK, envelope{"metadata": md, "movies": mvs})
}
