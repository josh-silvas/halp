package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func Test_setCredential(t *testing.T) {
	c, err := testCfg.setCredential(testCred.Username, testCred.Password, testSvc, testCred.Expire)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, c.Password, testCred.Password) {
		t.Fatalf("ERROR: assert.Equal %s", c.Password)
	}
	t.Logf("SUCCESS: created credential %s", c.Username)
}

func Test_getCredential(t *testing.T) {
	c, err := testCfg.getCredential(testCred.Username, testSvc)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, c.Password, testCred.Password) {
		t.Fatalf("ERROR: assert.Equal %s", c.Password)
	}
	t.Logf("SUCCESS: gathered credential %s, %d", c.Username, c.Expire)
}

func Test_isExpired(t *testing.T) {
	if expired := testCred.isExpired(); expired {
		t.Fatal("ERROR: incorrect Expire")
	}
	testCred.Expire = testCred.Expire - 1200
	if expired := testCred.isExpired(); !expired {
		t.Fatal("ERROR: incorrect Expire")
	}
	t.Logf("SUCCESS: credential is not expired")
}

func Test_deleteCredential(t *testing.T) {
	if err := keyring.Delete(testSvc.Name, testCred.Username); err != nil {
		t.Fatal(err)
	}
	t.Logf("SUCCESS: deleted credential %s", testSvc)
}
