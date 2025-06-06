package util

import "log"

// Recover logs recovered panic with context.
func Recover(context string) {
	if r := recover(); r != nil {
		log.Printf("%s panic recovered: %v", context, r)
	}
}
