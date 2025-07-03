package main

import (
	"testing"

)

func TestContentType(t *testing.T) {
	// Example 1: Content-Type negotiation
	accepts := []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"}
	offers := []string{"application/json", "text/html", "application/xml"}
	result := NegotiateContent(accepts, offers)
	if result != "text/html" {
		t.Errorf("Got wrong result: %v", result)
	}
}

func TestLanguage(t *testing.T) {
	// Example 2: Language negotiation with quality values
	accepts := []string{"en-US,en;q=0.9,fr;q=0.8,de;q=0.7,*;q=0.5"}
	offers := []string{"de", "fr", "es", "en"}
	result := NegotiateContent(accepts, offers)
	if result != "en" {
		t.Errorf("Got wrong result: %v", result)
	}
}

func TestEncoding(t *testing.T) {
	// Example 3: Encoding negotiation with wildcards
	accepts := []string{"gzip", "deflate", "br;q=0.9", "*;q=0.1"}
	offers := []string{"identity", "gzip", "brotli"}
	result := NegotiateContent(accepts, offers)
	if result != "gzip" {
		t.Errorf("Got wrong result: %v", result)
	}
}

func TestNoAcceptable(t *testing.T) {
	// Example 4: No acceptable content
	accepts := []string{"image/png", "image/jpeg"}
	offers := []string{"text/html", "application/json"}
	result := NegotiateContent(accepts, offers)
	if result != "" {
		t.Errorf("Negotiation should have failed")
	}
}

func TestNoHeader(t *testing.T) {
	// Example 5: No accept headers (should return first offer)
	accepts := []string{}
	offers := []string{"application/json", "text/html"}
	result := NegotiateContent(accepts, offers)
	if result != "application/json" {
		t.Errorf("Didn't get first offer")
	}
}

func TestWildcard(t *testing.T) {
	// Example 6: Complex media type negotiation
	accepts := []string{"application/*;q=0.9", "text/*;q=0.8", "*/*;q=0.1"}
	offers := []string{"image/png", "application/json", "text/plain", "text/html"}
	result := NegotiateContent(accepts, offers)
	if result != "application/json" {
		t.Errorf("wildcard mistakenly matched %v", result)
	}
}
