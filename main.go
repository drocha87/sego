package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type Lexer struct {
	content []rune
}

func NewLexer(s string) *Lexer {
	return &Lexer{content: []rune(s)}
}

func (l *Lexer) TrimLeft() {
	n := 0
	for n < len(l.content) && unicode.IsSpace(l.content[n]) {
		n++
	}
	l.content = l.content[n:]
}

func (l *Lexer) Chop(n int) []rune {
	result := make([]rune, n)
	for i, r := range l.content {
		result[i] = r
		if i == n-1 {
			break
		}
	}
	l.content = l.content[n:]
	return result
}

func (l *Lexer) ChopWhile(predicate func(rune) bool) []rune {
	n := 0
	for n < len(l.content) && predicate(l.content[n]) {
		n += 1
	}
	return l.Chop(n)
}

func (l *Lexer) IsEmpty() bool {
	return len(l.content) <= 0
}

func (l *Lexer) NextToken() (string, error) {
	l.TrimLeft()

	if l.IsEmpty() {
		return "", errors.New("EOF")
	}
	if unicode.IsNumber(l.content[0]) {
		return string(l.ChopWhile(unicode.IsNumber)), nil
	}
	if unicode.IsLetter(l.content[0]) {
		s := string(l.ChopWhile(func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }))
		return strings.ToUpper(s), nil
	}
	return string(l.Chop(1)), nil
}

//	fn tf(t: &str, d: &TermFreq) -> f32 {
//	    let a = d.get(t).cloned().unwrap_or(0) as f32;
//	    let b = d.iter().map(|(_, f)| *f).sum::<usize>() as f32;
//	    a / b
//	}
func tf(t string, d TermFreq) float32 {
	a := float32(d[t])
	var b float32
	for _, value := range d {
		b += float32(value)
	}
	return a / b
}

//	fn idf(t: &str, d: &TermFreqIndex) -> f32 {
//	    let N = d.len() as f32;
//	    let M = d.values().filter(|tf| tf.contains_key(t)).count().max(1) as f32;
//	    return (N / M).log10();
//	}
func idf(t string, d TermFreqIndex) float32 {
	N := float32(len(d))

	occurrences := float32(0)
	for _, tf := range d {
		if _, ok := tf[t]; ok {
			occurrences += 1
		}
	}
	M := float32(math.Max(float64(occurrences), 1))

	return float32(math.Log10(float64(N / M)))
}

func main() {
	err := entry()
	if err != nil {
		os.Exit(1)
	}
}

func usage(program string) {
	fmt.Printf("Usage: %s [SUBCOMMAND] [OPTIONS]\n", program)
	fmt.Println("Subcommands:")
	fmt.Println("    index <folder>                  index the <folder> and save the index to index.json file")
	fmt.Println("    search <index-file>             check how many documents are indexed in the file (searching is not implemented yet)")
	fmt.Println("    serve <index-file> [address]    start local HTTP server with Web Interface")
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

type TermFreq = map[string]uint64
type TermFreqIndex = map[string]TermFreq

func tfIndexOfFolder(dirPath string, tfIndex TermFreqIndex) error {
	fmt.Printf("Indexing directory %s...\n", dirPath)
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("ERROR: could not open directory %s for indexing: %s", dirPath, err.Error())
	}

	for _, file := range dir {
		filePath := path.Clean(dirPath + "/" + file.Name())
		fileType := file.Type()
		if fileType.IsDir() {
			err = tfIndexOfFolder(filePath, tfIndex)
			if err != nil {
				return err
			}
		} else {
			// TODO: how does this work with symlinks?
			fmt.Printf("Indexing %s...\n", filePath)
			content, err := parseEntireXmlFile(filePath)
			if err != nil {
				continue
			}
			tf := make(TermFreq)
			lexer := NewLexer(content)
			for {
				term, err := lexer.NextToken()
				if err != nil {
					break
				}
				tf[term] += 1
			}
			tfIndex[filePath] = tf
		}
	}
	return nil
}

func saveTfIndex(tfIndex TermFreqIndex, indexPath string) error {
	fmt.Printf("Saving %s...\n", indexPath)

	content, err := json.Marshal(tfIndex)
	if err != nil {
		return fmt.Errorf("ERROR: could not serialize index into file %s: %s", indexPath, err.Error())
	}

	// FIXME: WriteFile truncates it before writing, without changing permissions
	err = os.WriteFile(indexPath, content, 0666)
	if err != nil {
		return fmt.Errorf("ERROR: could not create index file %s: %s", indexPath, err.Error())
	}

	return nil
}

func entry() error {
	args := os.Args
	program := args[0]

	if len(args) < 2 {
		usage(program)
		return fmt.Errorf("ERROR: no subcommand was provided")
	}
	subcommand := args[1]

	switch subcommand {
	case "index":
		if len(args) < 3 {
			usage(program)
			return fmt.Errorf("ERROR: no directory is provided for %s subcommand", subcommand)
		}
		tfIndex := make(TermFreqIndex)
		err := tfIndexOfFolder(args[2], tfIndex)
		if err != nil {
			return err
		}
		err = saveTfIndex(tfIndex, "index.json")
		if err != nil {
			return err
		}

	case "search":
		if len(args) < 3 {
			usage(program)
			return fmt.Errorf("ERROR: no path to index is provided for %s subcommand", subcommand)
		}
		return checkIndex(args[2])

	case "serve":
		if len(args) < 3 {
			usage(program)
			return fmt.Errorf("ERROR: no path to index is provided for %s subcommand", subcommand)
		}
		indexPath := args[2]
		content, err := os.ReadFile(indexPath)
		if err != nil {
			return fmt.Errorf("ERROR: could not load %s: %s", indexPath, err)
		}
		var tfIndex TermFreqIndex
		err = json.Unmarshal(content, &tfIndex)
		if err != nil {
			return fmt.Errorf("ERROR: could not serialize file %s as json: %s", indexPath, err)
		}

		startServe(tfIndex)

	default:
		usage(program)
		return fmt.Errorf("ERROR: unknown subcommand %s", subcommand)
	}

	return nil
}

func checkIndex(indexPath string) error {
	fmt.Printf("Reading %s index file...\n", indexPath)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("ERROR: could not open index file %s: %s", indexPath, err.Error())
	}

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	if err != nil {
		return fmt.Errorf("ERROR: could not parse index file %s: %s", indexPath, err.Error())
	}
	fmt.Printf("%s contains %d files\n", indexPath, len(result))

	return nil
}
