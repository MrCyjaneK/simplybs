package pack

import "testing"

func TestApplyRootBuildOverrides(t *testing.T) {
	root := map[string]string{}
	applyRootBuildOverrides(root, 0)
	if root["FORCE_UNSAFE_CONFIGURE"] != "1" {
		t.Fatalf("uid 0: got %q", root["FORCE_UNSAFE_CONFIGURE"])
	}

	user := map[string]string{}
	applyRootBuildOverrides(user, 1000)
	if _, ok := user["FORCE_UNSAFE_CONFIGURE"]; ok {
		t.Fatalf("uid 1000 should not set FORCE_UNSAFE_CONFIGURE: %v", user)
	}
}
