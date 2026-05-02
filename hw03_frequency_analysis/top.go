package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}

	uniqueWords := make([]string, 0, len(counts))
	for word := range counts {
		uniqueWords = append(uniqueWords, word)
	}

	sort.Slice(uniqueWords, func(i, j int) bool {
		cntI, cntJ := counts[uniqueWords[i]], counts[uniqueWords[j]]
		if cntI != cntJ {
			return cntI > cntJ
		}
		return uniqueWords[i] < uniqueWords[j]
	})

	limit := 10
	if len(uniqueWords) < limit {
		limit = len(uniqueWords)
	}

	return uniqueWords[:limit]
}
