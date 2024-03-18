package common

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
)

// IsSolutionValid checks if first 4 bytes of hash(challenge+solution) is less than target.
func IsSolutionValid(challenge []byte, solutionBytes []byte, target uint32) (bool, error) {
	digest := sha256.New()
	_, err := digest.Write(challenge)
	if err != nil {
		return false, err
	}
	_, err = digest.Write(solutionBytes)
	if err != nil {
		return false, err
	}
	hash := digest.Sum(nil)

	return hashLessOrEqTarget(hash, target), nil
}

func Solve(ctx context.Context, challenge []byte, target uint32) ([]byte, error) {
	solutionBytes := make([]byte, 16)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		_, err := rand.Read(solutionBytes)
		if err != nil {
			return nil, err
		}

		if valid, err := IsSolutionValid(challenge, solutionBytes, target); err != nil {
			return nil, err
		} else if valid {
			break
		}
	}
	return solutionBytes, nil
}

func hashLessOrEqTarget(hash []byte, target uint32) bool {
	bytesToCompare := hash[:4]
	val := binary.BigEndian.Uint32(bytesToCompare)
	return val <= target
}
