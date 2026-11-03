package utils

import "uuid"

// UniqueID returns prefix followed by a UUIDv7.
//
// IDs are used as map keys, where a collision silently overwrites an entry
// (subscribers that never get events, terminals that replace one another).
// A bare timestamp collided on Windows' coarse clock; v7 is time-ordered but
// carries random bits, so IDs stay unique and still sort by creation time.
func UniqueID(prefix string) string {
	return prefix + uuid.NewV7().String()
}
