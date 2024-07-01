package main

import (
	"errors"
	"strings"
	"unicode/utf8"
)

func parsearg(args []string, handler func(rune, func() (string, error)) error) ([]string, error) {
	var arg string

	gets := func() (string, error) {
		t := ""
		if arg != "" {
			t, arg = arg, ""
			return t, nil
		} else if len(args) > 0 {
			t, args = args[0], args[1:]
			return t, nil
		}

		return "", errors.New("no more arguments")
	}

	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "-") {
			break
		}

		if args[0] == "--" {
			args = args[1:]
			break
		}

		arg, args = args[0], args[1:]
		arg = arg[1:]

		for len(arg) > 0 {
			p, size := utf8.DecodeRuneInString(arg)
			if p == utf8.RuneError {
				return nil, errors.New("invalid utf8")
			}
			arg = arg[size:]
			if err := handler(p, gets); err != nil {
				return nil, err
			}
		}
	}

	return args, nil
}
