package codec

import (
	"testing"
	"time"
)

func TestRoleIDAcceptsBase36AndRejectsEverythingElse(t *testing.T) {
	for _, id := range []string{"fi8c8111kow0", "0", "1y2p0ij32e8e7", "reader"} {
		if !ValidRoleID(id) {
			t.Fatalf("rejected a valid base-36 id: %q", id)
		}
	}
	// A hyphen, uppercase, a wildcard and anything past the int64 ceiling are
	// not base 36 — the ambiguity a generated id exists to remove.
	for _, id := range []string{"", "payslip-reader", "Payroll", "r*", "роль", "1y2p0ij32e8e77"} {
		if ValidRoleID(id) {
			t.Fatalf("accepted an id that is not a base-36 Snowflake: %q", id)
		}
	}
}

// A key slot is TEXT, so an unpadded revision would order "10" before "2". The
// padding is what makes ORDER BY key5 identical to numeric order.
func TestRevisionPaddingSortsNumerically(t *testing.T) {
	two, err := RenderRevision(2)
	if err != nil {
		t.Fatal(err)
	}
	ten, err := RenderRevision(10)
	if err != nil {
		t.Fatal(err)
	}
	if !(two < ten) {
		t.Fatalf("padded revisions sort wrongly: %q !< %q", two, ten)
	}
	if len(two) != RevisionWidth || len(ten) != RevisionWidth {
		t.Fatalf("width %d and %d, want %d", len(two), len(ten), RevisionWidth)
	}
}

func TestRevisionRoundTripsAndRejectsUnstorable(t *testing.T) {
	for _, revision := range []int64{1, 2, 10, 999999999} {
		slot, err := RenderRevision(revision)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ParseRevision(slot)
		if err != nil || got != revision {
			t.Fatalf("revision %d rendered %q parsed back as %d err=%v", revision, slot, got, err)
		}
	}
	for _, revision := range []int64{0, -1, 99999999999} {
		if _, err := RenderRevision(revision); err == nil {
			t.Fatalf("rendered a revision with no slot representation: %d", revision)
		}
	}
	// Padding is storage's, not the contract's: an unpadded slot is not a
	// revision this encoder wrote, so it is not trusted.
	for _, slot := range []string{"2", "0000000000", "00000000-1", ""} {
		if _, err := ParseRevision(slot); err == nil {
			t.Fatalf("parsed a slot this encoder never wrote: %q", slot)
		}
	}
}

// Two ids issued in the same millisecond must differ, which is the sequence's
// whole job, and every id must be a valid role id.
func TestSnowflakesIssueDistinctBase36IDs(t *testing.T) {
	frozen := time.UnixMilli(1788000000000)
	gen, err := NewSnowflakes(0, func() time.Time { return frozen })
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		id := gen.Next()
		if !ValidRoleID(id) {
			t.Fatalf("issued an id that is not a valid role id: %q", id)
		}
		if seen[id] {
			t.Fatalf("issued %q twice inside one millisecond", id)
		}
		seen[id] = true
	}
}

func TestSnowflakesRefuseANodeOutsideTheLayout(t *testing.T) {
	if _, err := NewSnowflakes(SnowflakeNodeMax+1, nil); err == nil {
		t.Fatal("accepted a node id the layout cannot hold")
	}
	if _, err := NewSnowflakes(-1, nil); err == nil {
		t.Fatal("accepted a negative node id")
	}
}

// Ids are time-sortable: a later millisecond yields a larger id, which is what
// makes them useful as a key.
func TestSnowflakesAreTimeOrdered(t *testing.T) {
	now := time.UnixMilli(1788000000000)
	gen, _ := NewSnowflakes(0, func() time.Time { return now })
	first := gen.Next()
	now = now.Add(time.Second)
	second := gen.Next()
	if !(len(first) < len(second) || (len(first) == len(second) && first < second)) {
		t.Fatalf("ids are not time-ordered: %q then %q", first, second)
	}
}
