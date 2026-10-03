package utils

import (
	"fmt"
	"sync/atomic"
	"time"
)

var idSeq atomic.Uint64

// UniqueID returns prefix followed by a value unique within this process.
//
// A timestamp alone is not enough: Windows' clock is far coarser than Linux's,
// so IDs taken in a tight loop collide there. Used as map keys, colliding IDs
// silently overwrite each other — subscribers that never get events and never
// shut down, terminals that replace one another. The counter guarantees
// uniqueness; the timestamp keeps IDs distinct across restarts and readable.
func UniqueID(prefix string) string {
	return fmt.Sprintf("%s%d-%d", prefix, time.Now().UnixNano(), idSeq.Add(1))
}
