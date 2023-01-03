package utils

import (
	"sort"
	"strings"
)

func MarshalKey(name string, labels []string) string {
	keyParam := make([]string, 0, len(labels)+1)
	keyParam = append(keyParam, name)
	keyParam = append(keyParam, labels...)

	sort.Strings(keyParam)

	key := strings.Join(keyParam, "_")

	return key
}

func GetLabels(mapLabels map[string]string) []string {
	labels := make([]string, 0, len(mapLabels))
	for k := range mapLabels {
		labels = append(labels, k)
	}

	return labels
}
