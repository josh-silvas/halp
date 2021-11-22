package shared

import (
	"strings"
	"testing"
)

func TestConcatVLANs(t *testing.T) {
	var (
		ints     = []int{2334, 2335, 101, 2224, 1997, 1998, 1999, 2000}
		expected = "101,1997-2000,2224,2334-2335"
	)

	vlanTest := ConcatVLANs(ints)
	if strings.Join(vlanTest, ",") != expected {
		t.Fatalf("ERROR: %s did not match expected value of %s", strings.Join(vlanTest, ","), expected)
	}
	t.Logf("SUCCESS: found expected value of %s", expected)
}
