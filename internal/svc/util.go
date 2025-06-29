package svc

import (
	"slices"
	"strings"
)

func dedup(kws []string) []string {
	slices.SortFunc(kws, strings.Compare)
	i, j := 0, 0
	for i < len(kws) {
		kws[j] = kws[i]
		j++
		for i < len(kws) && kws[i] == kws[j-1] {
			i++
		}
	}
	return kws[:j]
}
