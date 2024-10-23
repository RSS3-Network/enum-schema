package main

import (
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pascaldekloe/name"
)

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func trimValueNames(values []Value, prefix string) []Value {
	for i := range values {
		values[i].name = strings.TrimPrefix(values[i].name, prefix)
	}
	return values
}

// prefixValueNames adds a prefix to each name
func prefixValueNames(values []Value, prefix string) []Value {
	for i := range values {
		values[i].name = prefix + values[i].name
	}
	return values
}

func transformValueNames(values []Value, transformMethod string) []Value {
	var fn func(src string) string
	switch transformMethod {
	case "snake":
		fn = func(s string) string {
			return strings.ToLower(name.Delimit(s, '_'))
		}
	case "snake_upper", "snake-upper":
		fn = func(s string) string {
			return strings.ToUpper(name.Delimit(s, '_'))
		}
	case "kebab":
		fn = func(s string) string {
			return strings.ToLower(name.Delimit(s, '-'))
		}
	case "kebab_upper", "kebab-upper":
		fn = func(s string) string {
			return strings.ToUpper(name.Delimit(s, '-'))
		}
	case "upper":
		fn = func(s string) string {
			return strings.ToUpper(s)
		}
	case "lower":
		fn = func(s string) string {
			return strings.ToLower(s)
		}
	case "title":
		fn = func(s string) string {
			return strings.Title(s)
		}
	case "title-lower":
		fn = func(s string) string {
			title := []rune(strings.Title(s))
			title[0] = unicode.ToLower(title[0])
			return string(title)
		}
	case "first":
		fn = func(s string) string {
			r, _ := utf8.DecodeRuneInString(s)
			return string(r)
		}
	case "first_upper", "first-upper":
		fn = func(s string) string {
			r, _ := utf8.DecodeRuneInString(s)
			return strings.ToUpper(string(r))
		}
	case "first_lower", "first-lower":
		fn = func(s string) string {
			r, _ := utf8.DecodeRuneInString(s)
			return strings.ToLower(string(r))
		}
	case "whitespace":
		fn = func(s string) string {
			return strings.ToLower(name.Delimit(s, ' '))
		}
	default:
		fn = func(s string) string {
			return s
		}
	}

	for i, v := range values {
		after := fn(v.name)
		// If the original one was "" or the one before the transformation
		// was "" (most commonly if linecomment defines it as empty) we
		// do not care if it's empty.
		// But if any of them was not empty before then it means that
		// the transformed emptied the value
		if v.originalName != "" && v.name != "" && after == "" {
			log.Fatalf("transformation of %q (%s) got an empty result", v.name, v.originalName)
		}
		values[i].name = after
	}

	return values
}
