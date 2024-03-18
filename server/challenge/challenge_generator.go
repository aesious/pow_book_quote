package challenge

import (
	"crypto/rand"
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	challengeLen = 32
)

type Challenge [challengeLen]byte

func (ch Challenge) prefix() prefix {
	return prefix(ch[:prefLen])
}

func (ch Challenge) nonce() nonce {
	return nonce(ch[prefLen:])
}

func FromBytes(bytes []byte) (Challenge, error) {
	var resultChallenge Challenge
	if len(bytes) != challengeLen {
		return resultChallenge, errors.New("unexpected number of bytes for challenge")
	}
	return Challenge(bytes), nil
}

type nonce [16]byte

const (
	prefLen = 16

	// Max age of challenge is maxRecentPrefCount*prefUpdateFreqSec.
	maxRecentPrefCount = 2
	prefUpdateFreqSec  = 5
)

type prefix [prefLen]byte

type prefixWithSolvedNonces struct {
	pref prefix
	// set is here, so it can be easily expired with expiring prefix
	set sync.Map
}

var (
	// mtx guards recentPrefs
	mtx         sync.RWMutex
	recentPrefs = make([]*prefixWithSolvedNonces, 1, maxRecentPrefCount)
)

func init() {
	runChallengePrefixUpdate()
}

// runChallengePrefixUpdate updates current challenge prefix
// every few  seconds and keeps a few recent prefixes.
func runChallengePrefixUpdate() {
	recentPrefs[0] = newPrefixWithSolvedNonces()

	go func() {
		ticker := time.NewTicker(prefUpdateFreqSec * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			mtx.Lock()
			// Update prefix slice.
			if len(recentPrefs) == maxRecentPrefCount { // slice if full, expire oldest prefix
				// Expire prefix.
				copy(recentPrefs[1:], recentPrefs)
				recentPrefs[0] = newPrefixWithSolvedNonces()
			} else { // add new prefix in the beginning
				recentPrefs = append([]*prefixWithSolvedNonces{newPrefixWithSolvedNonces()}, recentPrefs...)
			}
			mtx.Unlock()
		}
	}()
}

func newPrefixWithSolvedNonces() *prefixWithSolvedNonces {
	newPrefix := &prefixWithSolvedNonces{}
	_, err := rand.Read(newPrefix.pref[:])
	if err != nil {
		log.WithError(err).Fatal("Could not generate new challenge prefix")
	}
	return newPrefix
}
