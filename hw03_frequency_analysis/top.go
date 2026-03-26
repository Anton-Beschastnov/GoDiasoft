package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	// Разделяем строку по любым пробельным символам
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	// Подсчитываем частоту появления каждого слова
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}

	// Формируем слайс из уникальных слов для последующей сортировки
	uniqueWords := make([]string, 0, len(counts))
	for word := range counts {
		uniqueWords = append(uniqueWords, word)
	}

	// Сортировка:
	// 1. По частоте (от большего к меньшему)
	// 2. Если частота равна — лексикографически (по алфавиту)
	sort.Slice(uniqueWords, func(i, j int) bool {
		cntI, cntJ := counts[uniqueWords[i]], counts[uniqueWords[j]]
		if cntI != cntJ {
			return cntI > cntJ
		}
		return uniqueWords[i] < uniqueWords[j]
	})

	// Возвращаем первые 10 слов (или меньше, если уникальных слов всего < 10)
	limit := 10
	if len(uniqueWords) < limit {
		limit = len(uniqueWords)
	}

	return uniqueWords[:limit]
}
