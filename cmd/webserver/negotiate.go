package main

import (
	"sort"
	"strconv"
	"strings"
)

// acceptValue represents a parsed accept header value
type acceptValue struct {
	value       string  // the media type or value
	quality     float64 // quality factor (0.0 to 1.0)
	specificity int     // specificity level for sorting
	order       int     // original order in the accept header
}

// NegotiateContent returns the best content to offer from a set of possible
// values, based on the preferences represented by the accept values.
func NegotiateContent(accepts []string, offers []string) string {
	if len(offers) == 0 {
		return ""
	}

	// If no accept values, return first offer
	if len(accepts) == 0 {
		return offers[0]
	}

	// Parse accept values (limit to first 32 to avoid DOS)
	acceptValues := parseAcceptValues(accepts)
	if len(acceptValues) > 32 {
		acceptValues = acceptValues[:32]
	}

	// Find best match
	bestMatch := ""
	bestScore := -1.0
	bestSpecificity := -1
	bestOrder := len(offers) // use offer order as tiebreaker

	for offerIdx, offer := range offers {
		score, specificity := calculateScore(offer, acceptValues)

		// Check if this is a better match
		if score > bestScore ||
			(score == bestScore && specificity > bestSpecificity) ||
			(score == bestScore && specificity == bestSpecificity && offerIdx < bestOrder) {
			bestMatch = offer
			bestScore = score
			bestSpecificity = specificity
			bestOrder = offerIdx
		}
	}

	if bestScore <= 0 {
		return ""
	}

	return bestMatch
}

// parseAcceptValues parses accept header values into structured format
func parseAcceptValues(accepts []string) []acceptValue {
	var values []acceptValue
	order := 0

	for _, accept := range accepts {
		// Split by comma to handle multiple values in one string
		parts := strings.Split(accept, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			value := acceptValue{
				quality: 1.0, // default quality
				order:   order,
			}

			// Split by semicolon to separate value from parameters
			sections := strings.Split(part, ";")
			value.value = strings.TrimSpace(sections[0])

			// Parse quality parameter
			for i := 1; i < len(sections); i++ {
				param := strings.TrimSpace(sections[i])
				if strings.HasPrefix(param, "q=") {
					if q, err := strconv.ParseFloat(param[2:], 64); err == nil {
						if q >= 0.0 && q <= 1.0 {
							value.quality = q
						}
					}
					break
				}
			}

			// Calculate specificity for media types
			value.specificity = calculateSpecificity(value.value)

			values = append(values, value)
			order++
		}
	}

	// Sort by quality (desc), then specificity (desc), then order (asc)
	sort.Slice(values, func(i, j int) bool {
		if values[i].quality != values[j].quality {
			return values[i].quality > values[j].quality
		}
		if values[i].specificity != values[j].specificity {
			return values[i].specificity > values[j].specificity
		}
		return values[i].order < values[j].order
	})

	return values
}

// calculateSpecificity returns specificity level for media type matching
func calculateSpecificity(mediaType string) int {
	if mediaType == "*/*" {
		return 0 // least specific
	}
	if strings.HasSuffix(mediaType, "/*") {
		return 1 // type wildcard
	}
	return 2 // exact match (most specific)
}

// calculateScore returns the quality score and specificity for an offer against accept values
func calculateScore(offer string, acceptValues []acceptValue) (float64, int) {
	for _, accept := range acceptValues {
		if matches(offer, accept.value) {
			return accept.quality, accept.specificity
		}
	}
	return 0.0, 0
}

// matches checks if an offer matches an accept value (including wildcards)
func matches(offer, accept string) bool {
	if accept == "*/*" {
		return true
	}

	if strings.HasSuffix(accept, "/*") {
		// Type wildcard (e.g., "text/*")
		offerType := strings.Split(offer, "/")[0]
		acceptType := strings.Split(accept, "/")[0]
		return offerType == acceptType
	}

	// Exact match
	return offer == accept
}
