package sqlite

import (
	"agentlabs.local/abv/domain"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// The catalog's own state, as an L1 record:
//
//	key1=abv  key2=catalog  key3=<application>
//	boundary=application, tenant_id=''
//
// `domain.Catalog` is the application's permission and scope catalog. Its
// permissions and scopes are already records of their own, so what is left is
// the catalog's two properties: the mode it validates in, and its version.
//
// The boundary is `application` rather than `tenant` because both are the
// application's, identical for every tenant — the same reason a permission
// definition sits there.
type catalogPayload struct {
	CompatibilityEnabled bool  `json:"compatibility_enabled"`
	// Generation is a counter, not a fact about the catalog anyone addresses. It
	// is here because a record store was the only place left to put it, and that
	// is worth saying plainly rather than dressing it up: it exists so a reader
	// can tell whether the catalog moved between two pages, and nothing reads it
	// for its own sake.
	//
	// Deriving it instead — COUNT(*) or MAX(ts) over the catalog's rows — is the
	// better answer and does not hold today: a status change rewrites a value
	// without changing the count, and `ts` has one-second resolution, so two
	// writes in the same second are indistinguishable. Revisit when `ts` earns a
	// finer stamp.
	Generation int64 `json:"generation"`
}

// readCatalogState returns the catalog's own state. An absent row is not "no
// such application" — that is the registry's answer, given before this runs. It
// is "no state recorded yet", and the defaults are the honest reading:
// relationship validation off under Q-041, generation zero. The row appears on
// the first catalog write.
func readCatalogState(ctx context.Context, conn *sql.Conn, applicationID string) (catalogPayload, error) {
	var raw string
	err := conn.QueryRowContext(ctx, `
		SELECT value FROM abv_l1_records
		 WHERE boundary='application' AND tenant_id='' AND key1='abv' AND key2='catalog' AND key3=?`,
		applicationID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return catalogPayload{}, nil
	}
	if err != nil {
		return catalogPayload{}, corruptOrDB(err)
	}
	var payload catalogPayload
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return catalogPayload{}, domain.ErrMalformed
	}
	return payload, nil
}

// bumpCatalogGeneration records that this application's catalog moved. It is
// called in the same transaction as the write it describes, which is what makes
// the read-page-reread guarantee true rather than nearly true.
//
// The row is created on first use, because a catalog that has never been written
// has nothing to version.
func bumpCatalogGeneration(ctx context.Context, conn *sql.Conn, applicationID string) error {
	current, err := readCatalogState(ctx, conn, applicationID)
	if err != nil {
		return err
	}
	current.Generation++
	return writeCatalogState(ctx, conn, applicationID, current)
}

// bumpEveryCatalogGeneration is the platform-permission case: a platform
// permission is in every application's catalog, so a write to it invalidates
// every cached view. One statement, as the column increment was.
//
// Applications whose catalog has never been written have no row and need none:
// a reader of an unwritten catalog has nothing cached to invalidate.
func bumpEveryCatalogGeneration(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(ctx, `
		UPDATE abv_l1_records
		   SET value = json_set(value, '$.generation', json_extract(value, '$.generation') + 1)
		 WHERE boundary='application' AND tenant_id='' AND key1='abv' AND key2='catalog'`)
	return classify(err)
}

func writeCatalogState(ctx context.Context, conn *sql.Conn, applicationID string, payload catalogPayload) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return domain.ErrMalformed
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records (boundary, tenant_id, key1, key2, key3, value)
		VALUES ('application', '', 'abv', 'catalog', ?, ?)
		ON CONFLICT (boundary, tenant_id, key1, key2, key3, key4, key5, key6, key7, key8, key9, key10)
		DO UPDATE SET value = excluded.value`,
		applicationID, string(raw))
	return classify(err)
}
