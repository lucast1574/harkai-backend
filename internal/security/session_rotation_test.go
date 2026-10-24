package security

import "testing"

func TestSessionRotationIsDeterministicAndDomainSeparated(t *testing.T) {
	access, refresh := DeriveSessionTokens("previous-secret", "rotation-one")
	againAccess, againRefresh := DeriveSessionTokens("previous-secret", "rotation-one")
	otherAccess, otherRefresh := DeriveSessionTokens("previous-secret", "rotation-two")
	if access != againAccess || refresh != againRefresh || access == refresh || len(access) != 43 || len(refresh) != 43 {
		t.Fatal("rotation must recover the same distinct 256-bit tokens")
	}
	if access == otherAccess || refresh == otherRefresh {
		t.Fatal("nonce did not change tokens")
	}
	changed, _ := DeriveSessionTokens("different-secret", "rotation-one")
	if changed == access {
		t.Fatal("previous secret did not affect tokens")
	}
}
