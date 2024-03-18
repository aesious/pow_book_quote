package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/Sergio471/pow/server/challenge"
	"github.com/Sergio471/pow/server/quotes"
)

const (
	defaultServerProcessingTimeout = 5 * time.Second
)

// Sends a new challenge to a requesting client.
func getChallengeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultServerProcessingTimeout)
	defer cancel()

	challengeBytes, err := challenge.NewChallenge(ctx)
	if err != nil {
		writeErrorToResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	challengeStr := base64.StdEncoding.EncodeToString(challengeBytes[:])
	_, err = fmt.Fprint(w, challengeStr)
	if err != nil {
		writeErrorToResponse(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// Checks if challenge and solution provided by a client are valid.
// If so, sends a quote from the book to the client.
func getBookQuoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultServerProcessingTimeout)
	defer cancel()

	challengeStr := r.FormValue("challenge")
	solutionStr := r.FormValue("solution")

	challengeBytes, err := base64.StdEncoding.DecodeString(challengeStr)
	if err != nil {
		writeErrorToResponse(w, "invalid base64-encoded value", http.StatusBadRequest)
		return
	}

	ch, err := challenge.FromBytes(challengeBytes)
	if err != nil {
		writeErrorToResponse(w, "challenge could not be constructed from provided bytes", http.StatusBadRequest)
		return
	}

	solutionBytes, err := base64.StdEncoding.DecodeString(solutionStr)
	if err != nil {
		writeErrorToResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = challenge.PassSolution(ctx, ch, solutionBytes)
	switch err {
	case challenge.ErrChallengeNotRecent:
		writeErrorToResponse(w, "challenge is too old", http.StatusBadRequest)
		return
	case challenge.ErrAlreadySolved:
		writeErrorToResponse(w, "challenge already solved", http.StatusBadRequest)
		return
	case challenge.ErrSolutionIsInvalid:
		writeErrorToResponse(w, "solution is invalid", http.StatusBadRequest)
		return
	}

	_, err = fmt.Fprint(w, quotes.GetRandomQuote())
	if err != nil {
		writeErrorToResponse(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func writeErrorToResponse(w http.ResponseWriter, errMsg string, code int) {
	w.WriteHeader(code)
	fmt.Fprint(w, errMsg)
}
