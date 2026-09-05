package main

import (
	"math"
	"sort"
	"unicode"
)

type SearchResult struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

type BM25Index struct {
	documents []Capability
	tokens    [][]string
	df        map[string]int
	avgLength float64
}

func NewBM25Index(capabilities []Capability) *BM25Index {
	index := &BM25Index{documents: append([]Capability{}, capabilities...), df: map[string]int{}}
	var total int
	for _, capability := range capabilities {
		tokens := tokenize(capability.Name + " " + capability.Summary)
		index.tokens = append(index.tokens, tokens)
		total += len(tokens)
		seen := map[string]bool{}
		for _, token := range tokens {
			if !seen[token] {
				index.df[token]++
				seen[token] = true
			}
		}
	}
	if len(capabilities) > 0 {
		index.avgLength = float64(total) / float64(len(capabilities))
	}
	return index
}

func (index *BM25Index) Query(query string, limit int) []SearchResult {
	if limit <= 0 || index.avgLength == 0 {
		return nil
	}
	queryTokens := tokenize(query)
	results := make([]SearchResult, 0, len(index.documents))
	for position, capability := range index.documents {
		frequencies := map[string]int{}
		for _, token := range index.tokens[position] {
			frequencies[token]++
		}
		var score float64
		for _, token := range queryTokens {
			frequency := frequencies[token]
			if frequency == 0 {
				continue
			}
			documentFrequency := index.df[token]
			idf := math.Log(1 + (float64(len(index.documents)-documentFrequency)+0.5)/(float64(documentFrequency)+0.5))
			denominator := float64(frequency) + 1.2*(1-0.75+0.75*float64(len(index.tokens[position]))/index.avgLength)
			score += idf * (float64(frequency) * 2.2 / denominator)
		}
		if score > 0 {
			results = append(results, SearchResult{ID: capability.ID, Score: score})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].ID < results[j].ID
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func tokenize(value string) []string {
	var result []string
	current := make([]rune, 0)
	flush := func() {
		if len(current) > 0 {
			result = append(result, string(current))
			current = current[:0]
		}
	}
	for _, character := range []rune(value) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			current = append(current, unicode.ToLower(character))
		} else {
			flush()
		}
	}
	flush()
	return result
}
