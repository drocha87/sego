package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
)

type Documents []*Document

func LoadDocumentsFromJson(filePath string) (Documents, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not load %s: %s", filePath, err)
	}
	var docs Documents
	err = json.Unmarshal(content, &docs)
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not serialize file %s as json: %s", filePath, err)
	}
	return docs, nil
}

func NewDocumentsFromFolder(folder string) (Documents, error) {
	var docs Documents

	var closure func(cwd string) error
	closure = func(cwd string) error {
		fmt.Printf("Indexing directory %s...\n", cwd)
		dir, err := os.ReadDir(cwd)
		if err != nil {
			return fmt.Errorf("ERROR: could not open directory %s for indexing: %s", cwd, err.Error())
		}

		for _, file := range dir {
			filePath := path.Clean(cwd + "/" + file.Name())
			fileType := file.Type()
			if fileType.IsDir() {
				err = closure(filePath)
				if err != nil {
					return err
				}
			} else {
				// TODO: how does this work with symlinks?
				doc, err := NewDocument(filePath)
				if err != nil {
					continue
				}
				docs = append(docs, doc)
			}
		}
		return nil
	}
	err := closure(folder)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (ds Documents) InverseDocumentFrequency(term string) float32 {
	N := len(ds)

	occurrences := 0
	for _, doc := range ds {
		if _, ok := doc.TermsFreq[term]; ok {
			occurrences += 1
		}
	}
	if occurrences == 0 {
		occurrences = 1
	}

	return float32(math.Log10(float64(N) / float64(occurrences)))
}

func (ds *Documents) SaveToJson(filePath string) error {
	fmt.Printf("Saving %s...\n", filePath)

	content, err := json.Marshal(ds)
	if err != nil {
		return fmt.Errorf("ERROR: could not serialize index into file %s: %s", filePath, err.Error())
	}

	// FIXME: WriteFile truncates it before writing, without changing permissions
	err = os.WriteFile(filePath, content, 0666)
	if err != nil {
		return fmt.Errorf("ERROR: could not create index file %s: %s", filePath, err.Error())
	}

	return nil
}
