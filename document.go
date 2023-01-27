package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type Document struct {
	Path           string
	TermsFreq      map[string]uint64
	TotalTermsFreq uint64
}

func NewDocument(documentPath string) (*Document, error) {
	fmt.Printf("Indexing %s...\n", documentPath)
	content, err := parseEntireXmlFile(documentPath)
	if err != nil {
		return nil, err
	}

	doc := &Document{Path: documentPath}
	doc.TermsFreq = make(map[string]uint64)

	lexer := NewLexer(content)
	for {
		term, err := lexer.NextToken()
		if err != nil {
			break
		}
		doc.TermsFreq[term] += 1
	}
	for _, freq := range doc.TermsFreq {
		doc.TotalTermsFreq += freq
	}

	return doc, nil
}

func (d *Document) TermFrequency(term string) float32 {
	return float32(d.TermsFreq[term]) / float32(d.TotalTermsFreq)
}

func parseEntireXmlFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("ERROR: could not read file %s: %s", filePath, err)
	}
	tokenizer := html.NewTokenizer(bytes.NewReader(content))
	var result strings.Builder
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			token := tokenizer.Token()
			s := token.String()
			s = strings.TrimSpace(s)
			if len(s) > 0 {
				result.WriteString(s + " ")
			}
		}
	}
	return result.String(), nil
}
