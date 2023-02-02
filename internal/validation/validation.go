package validation

import (
	"errors"
	"fmt"
	"log"
	"mdb/internal/data"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const minYear = 1899

func Errors(err error) map[string]string {
	var ve validator.ValidationErrors
	out := map[string]string{}
	var rte data.RuntimeErr
	switch {
	case errors.As(err, &rte):
		out["Runtime"] = fmt.Sprint(err)
	case errors.As(err, &ve):
		for _, v := range ve {
			msg := ""
			switch v.Tag() {
			case "required":
				msg = "Is needed"
			case "unique", "genre":
				msg = "Values should be unique"
			case "max":
				msg = "Should be less than " + v.Param()
			case "min":
				msg = "Should be greater than " + v.Param()
			case "oneof":
				msg = "Should be one of:" + v.Param()
			case "yearrange":
				msg = fmt.Sprintf("Should be less than or equal to %d and greater than %d", time.Now().Year(), minYear)
			default:
				msg = "Unknown error"
			}
			out[v.Field()] = msg
		}
	default:
		out["other"] = err.Error()
	}
	return out
}

func YearRange(fl validator.FieldLevel) bool {
	year := fl.Field().Int()
	if int(year) <= time.Now().Year() && year > minYear {
		return true
	}
	return false
}

func Genres(fl validator.FieldLevel) bool {
	// Underlying value is a slice of string, hence used interface method and then type casted to slice
	genres := fl.Field().Interface().([]string)
	// Len of genres is 1 although it is slice
	splitGenres := strings.Split(genres[0], ",")
	unique := make(map[string]string)
	log.Println(splitGenres)
	for _, v := range splitGenres {
		unique[v] = v
	}
	log.Println(unique)
	return len(splitGenres) == len(unique)
}

func OneOf(fl validator.FieldLevel) bool {
	match := strings.Split(fl.Param(), " ")
	userInput := fl.Field().String()
	for _, v := range match {
		if v == userInput {
			return true
		}
	}
	return false
}
