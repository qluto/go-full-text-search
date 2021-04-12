package searcher

import (
	_ "log"
	"testing"
)

func TestParseQuery(t *testing.T) {
	query := "foo AND bar"

	actual := parseQuery(query)
	expected := AndOperator{IndexReader{"foo"}, IndexReader{"bar"}}

	t.Skip("Skipping this test for now.")

	if actual != expected {
		t.Errorf("got: %v\nwant: %v", actual, expected)
	}

}
