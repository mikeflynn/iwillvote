package main

import (
	"fmt"
	"strings"
)

func Mapify(in string) map[string]string {
	if in == "" {
		return map[string]string{}
	}

	a := strings.Split(in, "&")
	m := make(map[string]string, len(a))
	for _, v := range a {
		x := strings.Split(v, "=")
		m[x[0]] = x[1]
	}

	return m
}

func Stringify(in map[string]string) string {
	if len(in) == 0 {
		return ""
	}

	out := []string{}
	for k, v := range in {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(out, "&")
}
