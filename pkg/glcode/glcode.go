// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package glcode

// Lookup returns the code description and a boolean representing if the code was found.
func Lookup(code string) (string, bool) {
	desc, ok := glCodes[code]
	return desc, ok
}
