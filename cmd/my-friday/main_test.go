package main

import (
	"errors"
	"testing"
)

func TestExitCategories(t *testing.T) {
	cases := []struct {
		args    []string
		message string
		code    int
	}{
		{[]string{"my-friday", "validate"}, "manifest schema: invalid", 6},
		{[]string{"my-friday", "recover"}, "journal mismatch", 5},
		{[]string{"my-friday", "init"}, "target is not empty", 3},
		{[]string{"my-friday", "init"}, "creation failed and was rolled back", 4},
		{[]string{"my-friday", "version", "extra"}, "usage: my-friday version", 2},
	}
	for _, tc := range cases {
		if got, _ := classifyError(tc.args, errors.New(tc.message)); got != tc.code {
			t.Errorf("%v: got %d want %d", tc.args, got, tc.code)
		}
	}
}
