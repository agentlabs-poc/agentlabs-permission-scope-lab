package hrms_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// goCommand prefers the toolchain that is running this test over whatever the
// PATH happens to hold, so the answer describes this build and not another one.
func goCommand(t *testing.T) string {
	t.Helper()
	if root := runtime.GOROOT(); root != "" {
		candidate := filepath.Join(root, "bin", "go")
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate
		}
	}
	found, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("no go toolchain to ask: %v", err)
	}
	return found
}

// The whole reason the client side was split out is that an application can link
// the gate without linking the authority domain: 879 lines instead of 13,674,
// and no reason to compile a record store to guard an endpoint.
//
// That claim was checked only by a demonstration script — which lives in a
// session folder, not in this repository, and runs when someone runs it. A
// reviewer showed what that is worth: two lines added to go.mod put
// agentlabs.local/abv into this module's dependency closure and every test in
// every module still passed.
//
// The closure is what matters, not the import list. apps/hrms imports neither
// domain directly today; the hole is a dependency pulled in through something
// it does import.
func TestTheApplicationLinksNoAuthorityDomain(t *testing.T) {
	listed, err := exec.Command(goCommand(t), "list", "-deps", "./...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps: %v\n%s", err, listed)
	}
	closure := strings.Fields(string(listed))
	forbidden, required := []string{}, map[string]bool{
		"agentlabs.local/authmiddleware": false,
		"agentlabs.local/authclient":     false,
	}
	for _, dependency := range closure {
		for _, domain := range []string{"agentlabs.local/abv", "agentlabs.local/registry"} {
			if dependency == domain || strings.HasPrefix(dependency, domain+"/") {
				forbidden = append(forbidden, dependency)
			}
		}
		if _, wanted := required[dependency]; wanted {
			required[dependency] = true
		}
	}
	if len(forbidden) != 0 {
		t.Fatalf("the application links the authority side: %v", forbidden)
	}
	// Without this the check would pass on an empty listing, a failed build, or
	// a module that links nothing at all — and would have been reassuring for
	// exactly as long as it was meaningless.
	for module, linked := range required {
		if !linked {
			t.Fatalf("%s is not in the dependency closure — this check proved nothing", module)
		}
	}
}
