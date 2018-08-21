package indexer

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	actual := tokenize("foo, bar buz.")

	expected := []string{"foo", "bar", "buz"}

	if actual == nil {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
	if len(actual) != len(expected) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("got: %v\nwant: %v", actual, expected)
		}
	}
}

func TestUniqCount(t *testing.T) {
	tokens := []string{"foo", "bar", "buz", "foo"}
	actual := uniqCount(tokens)

	expected := map[string]int{
		"foo": 2,
		"bar": 1,
		"buz": 1,
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}
}

func TestFilter(t *testing.T) {
	l := []Posting{{"docA", 1, 0.2}, {"docB", 2, 0.1}}
	target := "docA"
	filtered, isFiltered := filter(l, target)
	expected := []Posting{{"docB", 2, 0.1}}

	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("got: %v\nwant: %v", filtered, expected)
	}
	if !isFiltered {
		t.Errorf("got: %v\nwant: %v", isFiltered, true)
	}
}
