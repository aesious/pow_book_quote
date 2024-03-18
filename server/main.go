package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/Sergio471/pow/common"
	"github.com/Sergio471/pow/server/challenge"
	log "github.com/sirupsen/logrus"
)

const (
	port = 8081
)

var (
	difficultyTarget common.DifficultyTarget = 6000
)

func init() {
	flag.Var(&difficultyTarget, "difficultyTarget", "Target so that first 4 bytes of hash must be less than it")
	flag.Parse()
}

func main() {
	challenge.SetHashPuzzleDifficulty(difficultyTarget.UInt32())

	http.HandleFunc("/getChallenge", getChallengeHandler)
	http.HandleFunc("/getBookQuote", getBookQuoteHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	log.Printf("Difficulty target is %d", difficultyTarget)
	log.Printf("Listening on port %d", port)
	err := server.ListenAndServe()
	if err != nil {
		log.WithError(err).Fatal("Could not start the server")
	}
}
