package challenge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sergio471/pow/common"
	"github.com/stretchr/testify/assert"
)

func solve(ctx context.Context, t *testing.T, ch Challenge) []byte {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	solutionBytes, err := common.Solve(ctx, ch[:], targetForHashPuzzle)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatal("Consider increasing timeout")
	}
	assert.NoError(t, err)
	return solutionBytes
}

func TestSolveNewChallenge(t *testing.T) {
	ctx := context.Background()

	ch, err := NewChallenge(ctx)
	assert.NoError(t, err)

	solution := solve(ctx, t, ch)
	assert.NoError(t, PassSolution(ctx, ch, solution))
}

func TestInvalidSolution(t *testing.T) {
	ctx := context.Background()

	ch, err := NewChallenge(ctx)
	assert.NoError(t, err)

	solution := solve(ctx, t, ch)
	solution[0] ^= 1 // going to be flaky
	err = PassSolution(ctx, ch, solution)
	assert.ErrorIs(t, err, ErrSolutionIsInvalid)
}

func TestNotRecentChallenge(t *testing.T) {
	ctx := context.Background()

	ch := Challenge{1, 2, 3, 4} // zeroes at end

	solution := solve(ctx, t, ch)
	err := PassSolution(ctx, ch, solution)
	assert.ErrorIs(t, err, ErrChallengeNotRecent)
}

func TestAlreadySolvedChallenge(t *testing.T) {
	ctx := context.Background()

	ch, err := NewChallenge(ctx)
	assert.NoError(t, err)

	solution := solve(ctx, t, ch)
	assert.NoError(t, PassSolution(ctx, ch, solution))
	assert.ErrorIs(t, PassSolution(ctx, ch, solution), ErrAlreadySolved)
}
