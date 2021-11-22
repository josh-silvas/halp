package version

import (
	"strings"
	"testing"
	"time"
)

var (
	testCfgVer = CfgVer{
		Version:   SemVer("v1.2.0"),
		Timestamp: time.Now(),
	}
)

func TestCfgVer_String(t *testing.T) {
	verString := testCfgVer.String()
	if !strings.Contains(verString, testCfgVer.Version.String()) {
		t.Errorf("ToString Error %s", verString)
	}
	verStruct, err := Parse(verString)
	if err != nil {
		t.Fatal(err)
	}
	if verStruct.Version != testCfgVer.Version {
		t.Errorf("Parsing error %s != %s", verStruct.Version, testCfgVer.Version)
	}
	t.Logf("SUCCESS: CfgVer paresed: %s, Timestamp: %s", verStruct.Version, verStruct.Timestamp)
}

func TestFromAPI(t *testing.T) {
	apiVer, err := FromAPI()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SUCCESS: Found v%s as latest version", apiVer.String())
}
