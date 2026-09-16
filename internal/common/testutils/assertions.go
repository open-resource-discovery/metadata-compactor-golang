//go:build unit || integration

package testutils

import (
	"reflect"
	"testing"
)

func AssertDeepEquals(t *testing.T, expected any, found any) {
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("result does not match expected output\n  got:  %+v\n  want: %+v", found, expected)
	}
}

func AssertPanics(t *testing.T, message string) {
	t.Helper()

	if err := recover(); err == nil {
		t.Fatal(message)
	}
}
