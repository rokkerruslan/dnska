package testing

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Assert(t *testing.T, got, want interface{}) {
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
