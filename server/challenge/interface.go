package challenge

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"

	"github.com/Sergio471/pow/common"
)

// NewChallenge returns concatenation of the current challenge prefix and random nonce.
// Challenge prefix helps to check if challenge is generated recently.
// Challenge prefix is regularly updated by the server.
// Challenge prefix is better than timestamp because it is unpredictable
// and makes it impossible for an attacker to precalculate a valid challenge
// for some time in the future.
//
// Later, a client has to find a solution so that
// sha256(challenge + solution) < target.
func NewChallenge(ctx context.Context) (Challenge, error) {
	var resultChallenge Challenge

	// Fill prefix.
	mtx.RLock()
	copy(resultChallenge[:prefLen], recentPrefs[0].pref[:])
	mtx.RUnlock()

	// Theoretically, we could spend too much time locking-copy-ing-unlocking.
	if err := ctx.Err(); err != nil {
		return resultChallenge, err
	}

	// Add nonce.
	var nonce nonce
	_, err := rand.Read(nonce[:])
	if err != nil {
		return resultChallenge, err
	}
	copy(resultChallenge[prefLen:], nonce[:])

	return resultChallenge, nil
}

var (
	ErrChallengeNotRecent = errors.New("challenge is not recent")
	ErrAlreadySolved      = errors.New("challenge already solved")
	ErrSolutionIsInvalid  = errors.New("solution is invalid")
)

// PassSolution returns nil error iff challenge and solution are accepted.
// Otherwise, returns error describing why they were not accepted.
func PassSolution(ctx context.Context, ch Challenge, solutionBytes []byte) error {
	mtx.RLock()
	defer mtx.RUnlock()

	// Find corresponding recent prefix.
	posInRecent := -1
	pref := ch.prefix()
	for i, recentPref := range recentPrefs {
		if bytes.Equal(pref[:], recentPref.pref[:]) {
			// Found challenge in recent.
			posInRecent = i
			break
		}
	}
	if posInRecent == -1 {
		return ErrChallengeNotRecent
	}

	// Check if already solved.
	if _, ok := recentPrefs[posInRecent].set.Load(ch.nonce()); ok {
		return ErrAlreadySolved
	}

	if valid, err := common.IsSolutionValid(ch[:], solutionBytes, targetForHashPuzzle); err != nil {
		return err
	} else if !valid {
		return ErrSolutionIsInvalid
	}

	// Could be too much time spent above on cpu-intensive calculations.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Add to solved set if no other client managed to do it before.
	_, loaded := recentPrefs[posInRecent].set.LoadOrStore(ch.nonce(), struct{}{})
	if loaded {
		return ErrAlreadySolved
	}

	return nil
}
