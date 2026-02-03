package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type OrderItem struct {
	Name       string
	Price      float64
	Quantity   int
	Subtotal   float64
	Confidence float64
}

type OrderResult struct {
	Items  []OrderItem
	Errors []string
	Total  float64
}

type quantityMatch struct {
	quantity int
	item     string
}

var numberWords = map[string]int{
	"one": 1, "single": 1,
	"two": 2, "three": 3, "four": 4, "five": 5,
	"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	"eleven": 11, "twelve": 12, "dozen": 12,
}

func extractQuantities(text string) []quantityMatch {
	var matches []quantityMatch
	lower := strings.ToLower(text)

	// Split by "and" or commas to get individual items
	parts := regexp.MustCompile(`,\s*and\s+|,\s*|\s+and\s+`).Split(lower, -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for "a dozen X"
		if m := regexp.MustCompile(`^(a|an)\s+dozen\s+(.+)`).FindStringSubmatch(part); m != nil {
			matches = append(matches, quantityMatch{quantity: 12, item: strings.TrimSpace(m[2])})
			continue
		}

		// Check for digit quantities "3 tacos"
		if m := regexp.MustCompile(`^(\d+)\s+(.+)`).FindStringSubmatch(part); m != nil {
			qty, _ := strconv.Atoi(m[1])
			matches = append(matches, quantityMatch{quantity: qty, item: strings.TrimSpace(m[2])})
			continue
		}

		// Check for word quantities "two burritos"
		if m := regexp.MustCompile(`^(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|dozen|single)\s+(.+)`).FindStringSubmatch(part); m != nil {
			qty := numberWords[m[1]]
			matches = append(matches, quantityMatch{quantity: qty, item: strings.TrimSpace(m[2])})
			continue
		}

		// Check for "a/an X" (but not "a dozen")
		if m := regexp.MustCompile(`^(a|an)\s+(.+)`).FindStringSubmatch(part); m != nil {
			item := strings.TrimSpace(m[2])
			if !strings.HasPrefix(item, "dozen") {
				matches = append(matches, quantityMatch{quantity: 1, item: item})
			}
			continue
		}
	}

	return matches
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}

	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(b)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(a)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(a); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(b); i++ {
		for j := 1; j <= len(a); j++ {
			cost := 1

			if b[i-1] == a[j-1] {
				cost = 0
			}

			matrix[i][j] = min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)
		}
	}
	return matrix[len(b)][len(a)]
}

func calculateSimilarity(a, b string) float64 {
	longer, shorter := a, b

	if len(b) > len(a) {
		longer, shorter = b, a
	}

	if len(longer) == 0 {
		return 1.0
	}

	distance := levenshteinDistance(longer, shorter)
	return float64(len(longer)-distance) / float64(len(longer))
}

func normalizeWord(word string) string {
	return strings.TrimSuffix(word, "s")
}

func findBestMatch(userInput string, menu []MenuItem) (*MenuItem, float64) {
	input := strings.ToLower(strings.TrimSpace(userInput))
	normalizedInput := normalizeWord(input)

	var bestMatch *MenuItem
	var bestScore float64

	for i := range menu {
		item := &menu[i]
		itemName := strings.ToLower(item.Name)
		normalizedItemName := normalizeWord(itemName)

		if input == itemName || normalizedInput == normalizedItemName {
			return item, 1.0
		}

		inputWords := strings.Fields(input)
		itemWords := strings.Fields(itemName)

		allInputWordsMatch := true
		matchedWords := 0
		for _, iw := range inputWords {
			niw := normalizeWord(iw)
			wordFound := false
			for _, jw := range itemWords {
				njw := normalizeWord(jw)
				if jw == iw || njw == niw || calculateSimilarity(niw, njw) > 0.85 {
					wordFound = true
					matchedWords++
					break
				}
			}
			if !wordFound {
				allInputWordsMatch = false
			}
		}

		if allInputWordsMatch && matchedWords > 0 {
			score := float64(matchedWords) / float64(len(itemWords))
			score = score * 0.9
			if score > bestScore {
				bestMatch = item
				bestScore = score
			}
			continue
		}

		similarity := calculateSimilarity(normalizedInput, normalizedItemName)
		if similarity > bestScore && similarity > 0.7 {
			bestMatch = item
			bestScore = similarity
		}
	}
	return bestMatch, bestScore
}

func parseOrder(orderText string, menu []MenuItem) OrderResult {
	quantities := extractQuantities(orderText)
	result := OrderResult{}

	if len(quantities) == 0 {
		match, confidence := findBestMatch(orderText, menu)
		if match != nil {
			result.Items = append(result.Items, OrderItem{
				Name:       match.Name,
				Price:      match.Price,
				Quantity:   1,
				Subtotal:   match.Price,
				Confidence: confidence,
			})
			result.Total = match.Price
		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("Could not find a match for %q.", orderText))
		}
	} else {
		for _, q := range quantities {
			match, confidence := findBestMatch(q.item, menu)
			if match != nil {
				subtotal := match.Price * float64(q.quantity)
				result.Items = append(result.Items, OrderItem{
					Name:       match.Name,
					Price:      match.Price,
					Quantity:   q.quantity,
					Subtotal:   subtotal,
					Confidence: confidence,
				})
				result.Total += subtotal
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Could not find a match for %q.", q.item))
			}
		}
	}

	return result
}

func formatOrder(r OrderResult) string {
	var sb strings.Builder

	if len(r.Items) > 0 {
		sb.WriteString("=== YOUR ORDER ===\n")
		for _, item := range r.Items {
			sb.WriteString(fmt.Sprintf("%dx %s - $%.2f\n", item.Quantity, item.Name, item.Subtotal))
			if item.Confidence < 0.8 {
				sb.WriteString("(⚠️ Low confidence match - is this correct?)\n")
			}
		}
		sb.WriteString("---\n")
		sb.WriteString(fmt.Sprintf("TOTAL: $%.2f\n", r.Total))
	}

	if len(r.Errors) > 0 {
		sb.WriteString("\n=== ERRORS ===\n")
		for _, err := range r.Errors {
			sb.WriteString(fmt.Sprintf("❌ %s\n", err))
		}
	}

	return sb.String()
}
