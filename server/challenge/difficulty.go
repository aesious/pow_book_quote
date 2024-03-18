package challenge

import "sync"

var (
	targetForHashPuzzle uint32 = 6000 // default value
	once                sync.Once
)

// SetHashPuzzleDifficulty sets target for a solution, so that first 4 bytes of hash are less than target.
// Only the first call takes effect.
//
// On some regular domestic machines, a value of 1500 would likely lead to ~100ms calculation.
func SetHashPuzzleDifficulty(target uint32) {
	once.Do(func() {
		targetForHashPuzzle = target
	})
}
