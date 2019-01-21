// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// +build ignore

// Generates glcode.go
//
// This file downloads the OFM GL codes located on their website:
//  https://ofm.wa.gov/sites/default/files/public/legacy/policy/75.40.htm
// and creates Go values for each code with their name.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

var (
	downloadURL    = "https://ofm.wa.gov/sites/default/files/public/legacy/policy/75.40.htm"
	outputFilename = filepath.Join("pkg", "glcode", "lookup.go")
)

type code struct {
	number, description string
}

func main() {
	when := time.Now().Format("2006-01-02T03:04:05Z")
	who, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get user on %s", runtime.GOOS)
	}

	// Write copyright header
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// Generated on %s by %s, any modifications will be overwritten
package glcode
`, when, who.Username)

	// Download HTML
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Fatalf("error while downloading %s: %v", downloadURL, err)
	}
	defer resp.Body.Close()

	// Write GL codes to source code
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("ERROR: reading HTML response: %v", err)
	}

	// codeMap is the number -> code
	codeMap := make(map[string]code)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n == nil {
			return
		}
		var code code
		for i := range n.Attr {
			// The HTML looks like the following, so we need to grab each TD and merge their node data into a pair.
			//
			// <TR>
			//   <TD class="policynumber">0006</TD>
			//   <TD class="policytext">Estimated accrued receipts </TD>
			// </TR>
			if n.Attr[i].Key == "class" && n.Attr[i].Val == "policynumber" {
				if n.FirstChild == nil {
					continue
				}
				num := strings.TrimSpace(n.FirstChild.Data)
				switch num {
				case "", "u":
					continue // skip empty TD's and ones with <u>..</u>
				case "p":
					// Sometimes the code is wrapped in a <p>..</p> tag.
					// <TR>
					// 	<TD class="policynumber"><P>0005</P></TD>
					// 	<TD class="policytext"><P>Estimated unallotted FTEs </P></TD>
					// </TR>
					num = strings.TrimSpace(n.FirstChild.FirstChild.Data)
				}
				if num != "" && code.number == "" {
					code.number = num
				}

				// Grab associated metadata
				if n.NextSibling != nil && n.NextSibling.NextSibling != nil {
					data := strings.TrimSpace(n.NextSibling.NextSibling.FirstChild.Data)
					switch data {
					case "", "u":
						continue // skip empty TD's and ones with <u>..</u>
					case "p":
						// Grab if inside <p>..</p>
						data = strings.TrimSpace(n.NextSibling.NextSibling.FirstChild.FirstChild.Data)
					}

					// Remove extra UTF-8 characters that Go source rejects
					// From: https://stackoverflow.com/a/20403220
					if !utf8.ValidString(data) {
						v := make([]rune, 0, len(data))
						for i, r := range data {
							if r == utf8.RuneError {
								_, size := utf8.DecodeRuneInString(data[i:])
								if size == 1 {
									continue
								}
							}
							v = append(v, r)
						}
						data = string(v)
					}

					// Cleanup code description
					data = strings.NewReplacer("\n", "", "  ", "").Replace(data)
					if data == "" || data == "u" || data == "strong" || data == "table" {
						continue
					}
					if data != "" && code.description == "" {
						code.description = data
					}
				}
			}
		}
		if code.number != "" && code.description != "" {
			if _, exists := codeMap[code.number]; !exists {
				codeMap[code.number] = code
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	log.Printf("found %d codes\n", len(codeMap))

	var codes []code
	for _, code := range codeMap {
		codes = append(codes, code)
	}

	// sort codes and write to source file
	sort.Slice(codes, func(i, j int) bool {
		return codes[i].number < codes[j].number
	})
	fmt.Fprintln(&buf, "var glCodes = map[string]string{")
	for i := range codes {
		fmt.Fprintf(&buf, fmt.Sprintf(`"%s": "%s",`+"\n", codes[i].number, codes[i].description))
	}
	fmt.Fprintln(&buf, "}")

	// format source code and write file
	out, err := format.Source(buf.Bytes())
	if err != nil {
		ioutil.WriteFile(outputFilename, buf.Bytes(), 0644)
		log.Fatalf("error formatting output code, err=%v", err)
	}

	err = ioutil.WriteFile(outputFilename, out, 0644)
	if err != nil {
		log.Fatalf("error writing file, err=%v", err)
	}
}
