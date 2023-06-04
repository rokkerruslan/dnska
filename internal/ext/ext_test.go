package ext

import (
	"fmt"
	"testing"

	testing2 "github.com/rokkerruslan/dnska/testing"
)

func TestLongestSuffix(t *testing.T) {
	cases := []struct {
		a   string
		b   string
		out string
	}{
		{a: "", b: "", out: ""},
		{a: "a", b: "", out: ""},
		{a: "", b: "b", out: ""},
		{a: "abc", b: "bac", out: "c"},
		{a: "abcbb", b: "bacbb", out: "cbb"},
		{a: "abbcc", b: "abbcc", out: "abbcc"},
		{a: "bbcc", b: "abbcc", out: "bbcc"},
		{a: "bcc", b: "abbcc", out: "bcc"},
		{a: "abXcc", b: "abYcc", out: "cc"},
		{a: "abcXab", b: "abcYab", out: "ab"},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			testing2.Assert(t, longestSuffix(tc.a, tc.b), tc.out)
		})
	}
}
