package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"encoding/json/jsontext"
)

func ReplaceJSON(b []byte, kv map[string]any) ([]byte, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(b))
	out := new(bytes.Buffer)
	enc := jsontext.NewEncoder(out, jsontext.Multiline(true)) // expand for readability
	for {
		// Read a token from the input.
		tok, err := dec.ReadToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Check whether the token contains the string "Golang" and
		// replace each occurrence with "Go" instead.
		if t, ok := kv[string(dec.StackPointer())]; ok {
			switch v := t.(type) {
			case string:
				if tok.Kind() != '"' {
					return nil, errors.New("mismatch")
				}
				tok = jsontext.String(v)
			case float64:
				if tok.Kind() != '0' {
					return nil, errors.New("mismatch")
				}
				tok = jsontext.Float(v)
			case bool:
				if tok.Kind() != 't' || tok.Kind() != 'f' {
					return nil, errors.New("mismatch")
				}
				tok = jsontext.Bool(v)
			}
		}

		// Write the (possibly modified) token to the output.
		if err := enc.WriteToken(tok); err != nil {
			return nil, err
		}
	}

	return out.Bytes(), nil
}

func main() {
	// Example input with non-idiomatic use of "Golang" instead of "Go".
	const input = `{
		"title": "Golang version 1 is released",
		"author": "Andrew Gerrand",
		"date": "2012-03-28",
		"text": "Today marks a major milestone in the development of the Golang programming language.",
		"otherArticles": [
			"Twelve Years of Golang",
			"The Laws of Reflection",
			"Learn Golang from your browser"
		]
	}`

	// Using a Decoder and Encoder, we can parse through every token,
	// check and modify the token if necessary, and
	// write the token to the output.
	var replacements []jsontext.Pointer
	in := strings.NewReader(input)
	dec := jsontext.NewDecoder(in)
	out := new(bytes.Buffer)
	enc := jsontext.NewEncoder(out, jsontext.Multiline(true)) // expand for readability
	for {
		// Read a token from the input.
		tok, err := dec.ReadToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		// Check whether the token contains the string "Golang" and
		// replace each occurrence with "Go" instead.
		if tok.Kind() == '"' && strings.Contains(tok.String(), "Golang") {
			replacements = append(replacements, dec.StackPointer())
			tok = jsontext.String(strings.ReplaceAll(tok.String(), "Golang", "Go"))
		}

		// Write the (possibly modified) token to the output.
		if err := enc.WriteToken(tok); err != nil {
			log.Fatal(err)
		}
	}

	// Print the list of replacements and the adjusted JSON output.
	if len(replacements) > 0 {
		fmt.Println(`Replaced "Golang" with "Go" in:`)
		for _, where := range replacements {
			fmt.Println("\t" + where)
		}
		fmt.Println()
	}
	fmt.Println("Result:", out.String())

}
