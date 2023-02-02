package main

import (
	"mdb/internal/data"
	"strings"
)

type envelope map[string]any

func addDefaultValue(lm *data.ListMovie) {
	if lm.Genres == nil {
		lm.Genres = []string{}
	} else {
		lm.Genres = strings.Split(lm.Genres[0], ",")
	}
	if lm.Page == 0 {
		lm.Page = 1
	}
	if lm.PageSize == 0 {
		lm.PageSize = 2
	}
	if lm.Sort == "" {
		lm.Sort = "id"
	}
}
