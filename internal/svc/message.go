package svc

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrBadQuery = errors.New("bad query")
)

func Hash(kws []string) (string, error) {
	if len(kws) < 2 || len(kws) > 10 {
		return "", ErrBadQuery
	}
	slices.SortFunc(kws, strings.Compare)
	return "", nil
}
