package main

import (
	"encoding/json"
	"fmt"
	"os"
)

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

		docs, err := NewDocumentsFromFolder(args[2])
		if err != nil {
			return err
		}

		err = docs.SaveToJson("index.json")
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
		docs, err := LoadDocumentsFromJson(args[2])
		if err != nil {
			return err
		}
		startServe(docs)

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
