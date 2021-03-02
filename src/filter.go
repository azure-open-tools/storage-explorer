package main

import "strings"

type Filter struct {
	Key   string
	Value string
}

func containsMetadataMatch(metadata map[string]string, filter []Filter) bool {
	if len(metadata) == 0 {
		return false
	}

	for _, entry := range filter {
		if val, ok := metadata[entry.Key]; ok {
			if strings.Contains(val, entry.Value) {
				return true
			}
		}
	}
	return false
}

func createMetadataFilter(inputFilter []string) []Filter {
	var metadataFilter []Filter
	if len(inputFilter) > 0 {
		for _, entry := range largs.MetadataFilter {
			if strings.Contains(entry, ":") {
				split := strings.Split(entry, ":")
				f := Filter{split[0], split[1]}
				metadataFilter = append(metadataFilter, f)
			}
		}
	}
	return metadataFilter
}
