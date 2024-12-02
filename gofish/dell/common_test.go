/*
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dell

import (
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

func assertBool(t testing.TB, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

func assertInt(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func assertArray(t testing.TB, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
		return
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("got %v, want %v", got, want)
			return
		}
	}
}

func assertLinkArray(t testing.TB, got common.Links, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
		return
	}
	for i, v := range want {
		if string(got[i]) != v {
			t.Errorf("got %v, want %v", got, want)
			return
		}
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
