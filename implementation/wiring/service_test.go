package wiring_test

import (
	"agentlabs.local/wiring"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig(dir string) wiring.Config {
	return wiring.Config{
		AuthorityPath:  filepath.Join(dir, "authority.db"),
		RegistryPath:   filepath.Join(dir, "registry.db"),
		CreateRegistry: true,
		Administration: abvAdmin{}, RegistryAdministration: regAdmin{},
		Operator: operator, Clock: clock{},
	}
}

// A half-configured service is refused rather than assembled. Every field here
// is something only a deployment knows; defaulting one would mean guessing at
// who may administer, or which store to open.
func TestOpenRefusesIncompleteConfiguration(t *testing.T) {
	dir := t.TempDir()
	for name, spoil := range map[string]func(*wiring.Config){
		"no authority path": func(c *wiring.Config) { c.AuthorityPath = "" },
		"no registry path":  func(c *wiring.Config) { c.RegistryPath = "" },
		"no clock":          func(c *wiring.Config) { c.Clock = nil },
		"no administration": func(c *wiring.Config) { c.Administration = nil },
		"no registry admin": func(c *wiring.Config) { c.RegistryAdministration = nil },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := validConfig(dir)
			spoil(&cfg)
			service, err := wiring.Open(t.Context(), cfg)
			if err == nil {
				_ = service.Close()
				t.Fatal("assembled anyway")
			}
			if service != nil {
				t.Fatal("returned a service beside an error")
			}
		})
	}
	if _, err := wiring.Open(nil, validConfig(dir)); err == nil { //nolint:staticcheck // the nil context is the point
		t.Fatal("assembled without a context")
	}
}

// Opening leaves nothing behind when the second store fails: the registry is
// already open by then, and a leaked handle is the classic composition bug.
//
// Re-opening is not a test of this. SQLite lets a file be opened twice, so a
// leaked handle re-opens perfectly happily — an earlier version of this test
// passed with the cleanup deleted. Counting the descriptors is what actually
// fails when the handle survives.
func TestOpenClosesTheRegistryWhenAuthorityFails(t *testing.T) {
	dir := t.TempDir()
	cfg := validConfig(dir)
	// A directory where the authority store should be. The registry opens
	// first and must be closed again on the way out.
	if err := os.Mkdir(cfg.AuthorityPath, 0o755); err != nil {
		t.Fatal(err)
	}
	before := openHandles(t, cfg.RegistryPath)
	if service, err := wiring.Open(t.Context(), cfg); err == nil {
		_ = service.Close()
		t.Fatal("opened an authority store that is a directory")
	}
	if after := openHandles(t, cfg.RegistryPath); after > before {
		t.Fatalf("the failed attempt left %d registry handle(s) open", after-before)
	}
}

// openHandles counts this process's descriptors pointing at one file.
func openHandles(t *testing.T, path string) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skipf("descriptor counting is unavailable here: %v", err)
	}
	target, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		link, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err != nil {
			continue
		}
		if strings.HasPrefix(link, target) {
			count++
		}
	}
	return count
}

// The rule this package exists for, checked rather than asserted. It has been
// claimed in a doc comment, in a charter and in two pull requests; this is the
// first thing that can fail when it stops being true.
func TestNeitherDomainImportsTheOther(t *testing.T) {
	for _, pair := range []struct{ domain, forbidden string }{
		{"../abv", "agentlabs.local/registry"},
		{"../registry", "agentlabs.local/abv"},
	} {
		found, scanned := []string{}, 0
		err := filepath.WalkDir(pair.domain, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			scanned++
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(source), `"`+pair.forbidden) {
				found = append(found, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if scanned == 0 {
			t.Fatalf("scanned no Go files under %s — this check would pass for any rule", pair.domain)
		}
		if len(found) != 0 {
			t.Fatalf("%s imports %s in %v — the two domains must meet only here", pair.domain, pair.forbidden, found)
		}
	}
}
