package utils

import (
	"errors"
	"regexp"
)

var _prefix = regexp.MustCompile("^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$")
var _urlArg = regexp.MustCompile("^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$") // Yes, same for now.

// ErrInvalidPrefix tells the specified prefix is not of
// the valid format.
var ErrInvalidPrefix = errors.New("invalid prefix")

// ErrInvalidURLArg tells the specified URL arg is not of
// the valid format.
var ErrInvalidURLArg = errors.New("invalid URL arg")

// CheckPrefix checks a given prefix is valid.
func CheckPrefix(prefix string) error {
	if ok := _prefix.MatchString(prefix); !ok {
		return ErrInvalidPrefix
	}
	return nil
}

// CheckURLArg checks a given url arg (name) is valid.
func CheckURLArg(prefix string) error {
	if ok := _urlArg.MatchString(prefix); !ok {
		return ErrInvalidURLArg
	}
	return nil
}
