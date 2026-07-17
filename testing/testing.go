package testing

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Check compares the got and want values and reports an error if they are not equal.
func Check(t *testing.T, got, want any) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

func FailIfError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func ThisIsFine(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Error(err)
	}
}
