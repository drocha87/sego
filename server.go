package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func searchHandlerFactory(docs Documents) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		queries := r.URL.Query()
		query := queries["query"][0]

		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		type ResultPair struct {
			Path string  `json:"path,omitempty"`
			Freq float32 `json:"freq,omitempty"`
		}

		var result []ResultPair

		for _, doc := range docs {
			rank := float32(0)
			lexer := NewLexer(query)
			for {
				token, err := lexer.NextToken()
				if err != nil {
					break
				}
				rank += doc.TermFrequency(token) * docs.InverseDocumentFrequency(token)
			}
			if rank > 0 {
				result = append(result, ResultPair{Path: doc.Path, Freq: rank})
			}
		}

		w.Header().Set("Content-Type", "application/json")

		body, err := json.Marshal(map[string]interface{}{"length": len(result), "results": result})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not write response body: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func startServe(docs Documents) {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/search", searchHandlerFactory(docs))

	fmt.Printf("Starting server at port 6969\n")
	if err := http.ListenAndServe(":6969", nil); err != nil {
		log.Fatal(err)
	}
}
