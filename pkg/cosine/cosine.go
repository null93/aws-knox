package cosine

import (
	"math"
	"sort"
	"strings"
)

type result struct {
	Text       string
	Similarity float64
	SubString  bool
}

func charNGrams(text string, n int) map[string]float64 {
	text = strings.ToLower(strings.ReplaceAll(text, " ", "_"))
	counts := make(map[string]float64)
	for i := 0; i <= len(text)-n; i++ {
		gram := text[i : i+n]
		counts[gram]++
	}
	return counts
}

func similarity(a, b map[string]float64) float64 {
	var dotProduct, magA, magB float64
	for key, valA := range a {
		dotProduct += valA * b[key]
	}
	for _, val := range a {
		magA += val * val
	}
	for _, val := range b {
		magB += val * val
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(magA) * math.Sqrt(magB))
}

func Search(docs []string, query string, ngramSize int) []result {
	vector := charNGrams(query, ngramSize)
	results := []result{}
	for _, doc := range docs {
		isSubString := strings.Index(doc, query) != -1
		result := result{
			Text:       doc,
			Similarity: similarity(charNGrams(doc, ngramSize), vector),
			SubString:  isSubString,
		}
		if isSubString {
			result.Similarity += 0.25
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	return results
}
