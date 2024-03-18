package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Sergio471/pow/common"
	"github.com/sirupsen/logrus"
)

const (
	defaultReqTimeout = time.Second * 15
)

var (
	// Input params.
	host             string
	port             int
	difficultyTarget common.DifficultyTarget = 6000
	numRequests      int

	// Built from input params.
	getChallengeURL string
	getBookQuoteURL string
)

func init() {
	flag.Var(&difficultyTarget, "difficultyTarget", "Target so that first 4 bytes of hash must be less than it")
	flag.StringVar(&host, "host", "server", "Server host address")
	flag.IntVar(&port, "port", 8081, "Server port")
	flag.IntVar(&numRequests, "n", 10, "Number of requests to make (max 1000)")
	flag.Parse()

	if numRequests > 1000 {
		logrus.Fatal("number of requests can not exceed 1000")
	}

	getChallengeURL = fmt.Sprintf("http://%s:%d/getChallenge", host, port)
	getBookQuoteURL = fmt.Sprintf("http://%s:%d/getBookQuote", host, port)
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, defaultReqTimeout*time.Duration(numRequests)+time.Second*10)
	defer cancel()

	startTime := time.Now()
	var wg sync.WaitGroup
	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		reqId := i
		// TODO: make worker pool of goroutines and allow more than 1000 requests
		go func() {
			defer wg.Done()
			retrieveBookQuote(ctx, reqId)
		}()
	}
	wg.Wait()
	duration := time.Since(startTime)

	logrus.Printf("Total time spent: %s\n", duration)
}

func retrieveBookQuote(ctx context.Context, reqId int) {
	ctx, cancel := context.WithTimeout(ctx, defaultReqTimeout)
	defer cancel()

	log := logrus.WithFields(logrus.Fields{
		"reqId": reqId,
	})
	log.Println("Retrieving book quote...")
	startTime := time.Now()
	defer func() {
		log.Printf("Time spent: %s\n", time.Since(startTime))
	}()

	challenge, err := getChallenge(ctx, getChallengeURL)
	if err != nil {
		log.Errorf("Could not get challenge from server: %s\n", err)
		return
	}
	log.Printf("Received base64-encoded challenge: %s\n", challenge)

	solution, err := findSolution(ctx, challenge)
	if err != nil {
		log.Errorf("Could not calculate a solution for the challenge: %s\n", err)
		return
	}

	bookQuote, err := getBookQuote(ctx, challenge, solution)
	if err != nil {
		log.Errorf("Could not get a book quote from server: %s\n", err)
		return
	}
	log.Printf("Got a book quote: \"%s\"\n", bookQuote)
}

func findSolution(ctx context.Context, challengeStr string) (string, error) {
	challengeBytes, err := base64.StdEncoding.DecodeString(challengeStr)
	if err != nil {
		return "", err
	}

	solutionBytes, err := common.Solve(ctx, challengeBytes, difficultyTarget.UInt32())
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(solutionBytes[:]), nil
}

var client = http.Client{
	Timeout: 5 * time.Second,
}

func getChallenge(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func getBookQuote(ctx context.Context, challenge, solution string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getBookQuoteURL, nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Set("challenge", challenge)
	q.Set("solution", solution)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status code: %d, message: %s", resp.StatusCode, bodyBytes)
	}

	return string(bodyBytes), nil
}
