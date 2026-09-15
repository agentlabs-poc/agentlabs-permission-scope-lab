package wiring_test

import (
	regdomain "agentlabs.local/registry/domain"
	"agentlabs.local/wiring"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig(dir string) wiring.Config {
	return wiring.Config{
		AuthorityPath:  filepath.Join(dir, "authority.db"),
		RegistryPath:   filepath.Join(dir, "registry.db"),
		CreateRegistry: true, CreateAuthority: true,
		Administration: abvAdmin{}, RegistryAdministration: regAdmin{},
		Operator: operator, Clock: clock{},
	}
}

// A half-configured service is refused rather than assembled, and refused *here*
// rather than somewhere downstream. Asserting only that an error came back was
// vacuous: registry.Open and abv.OpenSQLite already fail on every spoiled
// config, so the test passed with both guards deleted. It now names the message
// this package owns, and checks that nothing was opened on the way to it.
func TestOpenRefusesIncompleteConfiguration(t *testing.T) {
	for name, spoil := range map[string]func(*wiring.Config){
		"no authority path": func(c *wiring.Config) { c.AuthorityPath = "" },
		"no registry path":  func(c *wiring.Config) { c.RegistryPath = "" },
		"no clock":          func(c *wiring.Config) { c.Clock = nil },
		"no administration": func(c *wiring.Config) { c.Administration = nil },
		"no registry admin": func(c *wiring.Config) { c.RegistryAdministration = nil },
	} {
		t.Run(name, func(t *testing.T) {
			// A directory of its own: a guard that fires late leaves a store
			// behind, and sharing one would hide that.
			dir := t.TempDir()
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
			if !strings.Contains(err.Error(), "required") {
				t.Fatalf("refused downstream rather than here: %v", err)
			}
			entries, readErr := os.ReadDir(dir)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(entries) != 0 {
				t.Fatalf("a refused configuration still opened %d file(s)", len(entries))
			}
		})
	}
	if _, err := wiring.Open(nil, validConfig(t.TempDir())); err == nil { //nolint:staticcheck // the nil context is the point
		t.Fatal("assembled without a context")
	}
}

// The port is built after the registry is open, so its failure is the second
// leak path — the same class as the one below, and it had no test. A zero
// Operator is what makes NewPort refuse.
func TestOpenClosesTheRegistryWhenThePortFails(t *testing.T) {
	dir := t.TempDir()
	cfg := validConfig(dir)
	cfg.Operator = regdomain.Identity{}
	before := openHandles(t, cfg.RegistryPath)
	if service, err := wiring.Open(t.Context(), cfg); err == nil {
		_ = service.Close()
		t.Fatal("assembled with an operator the port refuses")
	}
	if after := openHandles(t, cfg.RegistryPath); after > before {
		t.Fatalf("the failed port left %d registry handle(s) open", after-before)
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
			// Parsed, not grepped. These packages write prose doc comments about
			// exactly this boundary, and a comment naming the other module's
			// path would fail a text search — a check that cries wolf gets
			// deleted the first time it is wrong.
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imported := range file.Imports {
				if strings.HasPrefix(strings.Trim(imported.Path.Value, `"`), pair.forbidden) {
					found = append(found, path)
				}
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

// A missing authority store is refused rather than created. A typo'd path used
// to succeed and leave an empty database behind, and an empty tenant answers
// every question with "no" — a deny-everything service instead of an error an
// operator can see.
func TestOpenRefusesAMissingAuthorityStore(t *testing.T) {
	dir := t.TempDir()
	cfg := validConfig(dir)
	cfg.CreateAuthority = false
	service, err := wiring.Open(t.Context(), cfg)
	if err == nil {
		_ = service.Close()
		t.Fatal("created the authority store behind the caller's back")
	}
	if _, statErr := os.Stat(cfg.AuthorityPath); statErr == nil {
		t.Fatal("a store was left on disk by the refused attempt")
	}
}

// A typed nil in an interface is not nil. The registry half is the load-bearing
// one: without the reflective check, wiring.Open accepts a typed-nil registry
// administration and returns a working-looking Service whose gates nil-deref on
// the first request. The authority half is caught downstream by mutation.New
// anyway, and is kept because a reader should not have to know which is which.
func TestOpenRefusesATypedNilAdministration(t *testing.T) {
	dir := t.TempDir()
	for name, spoil := range map[string]func(*wiring.Config){
		"authority administration": func(c *wiring.Config) { var typed *abvAdmin; c.Administration = typed },
		"registry administration":  func(c *wiring.Config) { var typed *regAdmin; c.RegistryAdministration = typed },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := validConfig(dir)
			spoil(&cfg)
			if service, err := wiring.Open(t.Context(), cfg); err == nil {
				_ = service.Close()
				t.Fatal("assembled with a typed-nil administration")
			}
		})
	}
}
