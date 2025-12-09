package utils

import (
	"unicode/utf8"

	"github.com/flosch/pongo2/v6"
)

func init() {
    pongo2.RegisterFilter("truncatechars", filterTruncateChars)
    // Add other missing filters here if needed
}

func filterTruncateChars(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    s := in.String()
    n := param.Integer()
    
    if utf8.RuneCountInString(s) <= n {
        return in, nil
    }
    
    // Naive truncation (bytes vs runes, but good enough for demo)
    // Proper way: iterate runes
    runes := []rune(s)
    if len(runes) > n {
        return pongo2.AsValue(string(runes[:n]) + "..."), nil
    }
    return in, nil
}
