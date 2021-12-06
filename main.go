package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	flag "github.com/jessevdk/go-flags"
)

func main() {
	var opts options
	_, err := flag.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	inputs, err := ParseOpts(opts)
	if err != nil {
		log.Fatal(err)
	}

	// If we've managed to parse the inputs we also want
	// to check if the user might be overwriting the file
	// and if they're alright with it
	if exists(inputs.OutputFp) {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Would you like to overwrite the file? (y/N): ")
		scanner.Scan()
		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
			log.Fatal(errors.New("file already exists"))
		}
	}

	c, err := inputs.Command()
	if err != nil {
		log.Fatal(err)
	}

	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func exists(fp string) bool {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return false
	}
	return true
}