package util

import "strings"

func ContainsStringFromSlice(str string, slice []string) bool {
	for _, s := range slice {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}
