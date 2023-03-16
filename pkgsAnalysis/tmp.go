package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	response, err := http.Get("https://index.golang.org/index")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// grep ImportedBy-detailsIndent from response.Body
	// and write to standard output
	// read one line from response.Body
	scanner := bufio.NewScanner(response.Body)
	type Message struct {
		Path, Version, Timestamp string
	}
	for scanner.Scan() {
		// json dump

		dec := json.NewDecoder(strings.NewReader(scanner.Text()))

		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", m.Path)
	}

}
