package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func getChars(charSet string) string {
	switch charSet {
	case "list":
		return "all bytes alphanumeric alphanumeric-nosim numeric alphabet binary hexadecimal"
	case "numeric":
		return "0123456789"
	case "bin":
		fallthrough
	case "binary":
		return "01"
	case "hex":
		fallthrough
	case "hexadecimal":
		return "0123456789ABCDEF"
	case "alphanumeric-nosim":
		fallthrough
	case "alpha-nosim":
		return "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnopqrstuvwxyz"
	case "alphabet":
		return "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnopqrstuvwxyz"
	case "alphanumeric":
		fallthrough
	case "alpha":
		return "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	default:
		// Everything
		return "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_-!$%^&();:.,<>/?"
	}
}

// Return a randomish string of the specified size,
// using the global CharSet
func randString(size int, charset string) string {
	bytes := make([]byte, size)
	setLen := byte(len(charset))

	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = charset[v%setLen]
	}
	return string(bytes)
}

func randBytes(size int) []byte {
	bytes := make([]byte, size)

	rand.Read(bytes)
	return bytes
}

func blockstring(s string, n int) string {
	var buffer bytes.Buffer
	var n1 = n - 1
	var l1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n1 && i != l1 {
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}

// Unique is a simple deduplicating cache
type Unique map[string]bool

func (u Unique) isUnique(str string) bool {
	if _, ok := u[str]; ok {
		// exists
		return false
	}
	// doesn't exist
	u[str] = true
	return true
}

func main() {

	var (
		charset      string
		stringlength int
		stringcount  int
		mangle       string
		b64          bool
		block        bool
		keyblock     bool
		pin          int
		blocksize    int
		unique       bool
		separator    string
		custom       string

		chars string
		uniq  Unique = make(Unique)
	)

	flag.StringVar(&charset, "chars", "all", "Characters to use ("+getChars("list")+")")
	flag.IntVar(&stringlength, "length", 20, "Length of string")
	flag.IntVar(&stringcount, "count", 20, "Number of strings")
	flag.StringVar(&mangle, "mangle", "", "Mangle the output (WARN: Decreases cardinality, should not be used with --base64) (UC LC)")
	flag.BoolVar(&b64, "base64", false, "Base64 encode the output")
	flag.BoolVar(&block, "block", false, "Block the output to 65 character lines")
	flag.BoolVar(&keyblock, "keyblock", false, "Shortcut to '--char bytes --base64 --block --blocksize 65' (HINT: --length 741, perhaps?)")
	flag.IntVar(&pin, "pin", 0, "Shortcut to '--char numeric --length int'")
	flag.IntVar(&blocksize, "blocksize", 65, "Slight misnomer: if --block is used, sets the line length to int")
	flag.BoolVar(&unique, "unique", false, "Ensure generated strings are unique. Lame")
	flag.StringVar(&separator, "separator", "\n", "What character or string should each value be separated with?")
	flag.StringVar(&custom, "custom", "", "A list of characters you want to use in lieu of '--chars' (repeat for prevalence)")
	flag.Parse()

	// Sanity
	if charset == "list" {
		charset = "all"
	}
	if keyblock {
		charset = "bytes"
		b64 = true
		block = true
		blocksize = 65
	}
	if pin > 0 {
		charset = "numeric"
		stringlength = pin
	}

	// String Quoting Madness
	if separator == "\n" {
		separator = `"\n"`
	} else {
		separator = `"` + separator + `"`
	}
	separator, err := strconv.Unquote(separator)
	if err != nil {
		fmt.Printf("Separator error: %s\n", err.Error())
		return
	}

	// Yes, globals suck, and there are "better" ways to do this.
	// No, in this instance it doesn't matter. Short-lived program, and it spares
	//   us a ton of computation by doing this.
	if custom != "" {
		chars = custom
	} else if charset != "bytes" {
		chars = getChars(charset)
	}

	// Print All The Strings!
	for i := 0; i < stringcount; i++ {

		var s string
		if charset != "bytes" {
			// Strings!
			s = randString(stringlength, chars)

			if b64 {
				s = base64.StdEncoding.EncodeToString([]byte(s))
			}

		} else {
			// Bytes!
			b := randBytes(stringlength)

			if b64 {
				s = base64.StdEncoding.EncodeToString(b)
			} else {
				s = string(b)
			}
		}
		// POST: s is a populated string of something

		switch strings.ToLower(mangle) {
		case "uc":
			s = strings.ToUpper(s)
		case "lc":
			s = strings.ToLower(s)
		}

		if block {
			s = blockstring(s, blocksize)
		}

		if unique && !uniq.isUnique(s) {
			continue
		}

		fmt.Printf("%s%s", s, separator)
	}
}
