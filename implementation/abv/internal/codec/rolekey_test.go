package codec

import "testing"

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
