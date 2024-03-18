package quotes

import (
	_ "embed"
	"math/rand"
	"strings"
)

var (
	//go:embed quotes.txt
	quotesFileContent string

	quotes []string
)

func init() {
	quotes = strings.Split(quotesFileContent, "\n")
}

func GetRandomQuote() string {
	idx := rand.Int() % len(quotes)
	return quotes[idx]
}
