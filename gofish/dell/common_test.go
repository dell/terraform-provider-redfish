package dell

import (
	"reflect"
	"testing"

	"github.com/stmcginnis/gofish/common"
)

// This file contains useful functions for testing purposes

func assertField(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func assertArray(t testing.TB, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertLinkArray(t testing.TB, got common.Links, want []string) {
	t.Helper()
	var linksToAssert common.Links
	for _, v := range want {
		linksToAssert = append(linksToAssert, common.Link(v))
	}
	if !reflect.DeepEqual(got, linksToAssert) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertLink(t testing.TB, got common.Link, want string) {
	t.Helper()
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func assertMapKeyValue(t testing.TB, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
