package lhdiff

import (
	v "github.com/rexsimiloluwah/distance_metrics/vector"
	"math"
	"strings"
)

func TfIdfCosineSimilarity(docA string, docB string) float64 {
	tokensA := strings.Fields(docA)
	tokensB := strings.Fields(docB)

	tokens := union(tokensA, tokensB)
	n := len(tokens)
	vectorA := make([]float64, n)
	vectorB := make([]float64, n)

	var (
		allTokens []string
	)
	allTokens = append(allTokens, tokensA...)
	allTokens = append(allTokens, tokensB...)
	if len(allTokens) == 0 {
		return 1
	}
	var documentFrequency = map[string]int{}
	for _, token := range allTokens {
		if documentFrequency[token] == 0 {
			documentFrequency[token] = 1
		} else {
			documentFrequency[token] = documentFrequency[token] + 1
		}
	}

	for k, token := range tokens {
		vectorA[k] = tfidf(token, tokensA, n, documentFrequency)
		vectorB[k] = tfidf(token, tokensB, n, documentFrequency)
	}

	similarity := v.CosineSimilarity(vectorA, vectorB)
	return similarity
}

func union(a, b []string) []string {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

func tfidf(token string, tokens []string, n int, documentFrequency map[string]int) float64 {
	tf := float64(count(token, tokens)) / float64(documentFrequency[token])
	idf := math.Log(float64(n) / (float64(documentFrequency[token])))
	return tf * idf
}

func count(key string, a []string) int {
	count := 0
	for _, s := range a {
		if key == s {
			count = count + 1
		}
	}
	return count
}
