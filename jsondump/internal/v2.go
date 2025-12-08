package internal

import (
	"bytes"
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
	// Keep track of the field-value...
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
		key := strings.TrimPrefix(string(dec.StackPointer()), "/") // Last token doesn't have the initial slash before the key.
		k, _ := dec.StackIndex(dec.StackDepth())
		fmt.Println(key, dec.StackDepth(), string(k))
		if v, ok := kv[key]; ok {
			// Write the key
			if k, _ := dec.StackIndex(dec.StackDepth()); k == '[' {
				// replace array item directly.
				switch tok.Kind() {
				case '"':
					tok = jsontext.String(v.(string))
				case '0':
					tok = jsontext.Float(v.(float64))
				case 't', 'f':
					tok = jsontext.Bool(v.(bool))
				}
				delete(kv, key)
			} else {
				if err := enc.WriteToken(tok); err != nil {
					return nil, err
				}
				delete(kv, key)
				// Read the value
				tok, err = dec.ReadToken()
				if err != nil {
					return nil, err
				}

				switch tok.Kind() {
				case '"':
					tok = jsontext.String(v.(string))
				case '0':
					tok = jsontext.Float(v.(float64))
				case 't', 'f':
					tok = jsontext.Bool(v.(bool))
				}
			}
		}

		// Write the (possibly modified) token to the output.
		if err := enc.WriteToken(tok); err != nil {
			return nil, err
		}
	}
	if len(kv) > 0 {
		panic(fmt.Sprint(kv))
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
