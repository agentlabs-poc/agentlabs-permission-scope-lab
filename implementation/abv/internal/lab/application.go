package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	labScenario       = "team-fin-c17"
	labFixtureContext = "maya-team1"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

type labApplication struct {
	area   domain.Area
	path   string
	facade *abv.Facade
}

var _ application.API = (*labApplication)(nil)
var _ application.Connect = Connect

// Connect opens an existing ABV database. It never creates or repairs one.
func Connect(ctx context.Context, area domain.Area, path string) (application.API, func() error, error) {
	if err := area.Validate(); err != nil {
		return nil, nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, nil, errors.Join(domain.ErrUnavailable, err)
	}
	fixture := TeamFINC17(area)
	administration, err := NewGrantStatusAdministration(area, fixture.Administration)
	if err != nil {
		return nil, nil, err
	}
	facade, err := abv.OpenSQLite(ctx, path, administration, clock{})
	if err != nil {
		return nil, nil, err
	}
	app := &labApplication{area: area, path: path, facade: facade}
	return app, facade.Close, nil
}

func (a *labApplication) Inspect(ctx context.Context, area domain.Area, kind, id string) (domain.Record, error) {
	if area != a.area {
		return domain.Record{}, domain.ErrRejected
	}
	return a.facade.Inspect(ctx, area, kind, id)
}

func (a *labApplication) CheckAssignment(ctx context.Context, area domain.Area, raw []byte) (domain.Diagnostic, error) {
	if area != a.area {
		return domain.Diagnostic{}, domain.ErrRejected
	}
	return a.facade.CheckAssignment(ctx, area, raw)
}

func (a *labApplication) Assign(ctx context.Context, area domain.Area, fixtureContext domain.FixtureContext, raw []byte) (domain.Receipt, error) {
	if area != a.area || fixtureContext.Name != labFixtureContext {
		return domain.Receipt{}, domain.ErrRejected
	}
	if err := verifyMarker(ctx, a.path, area); err != nil {
		return domain.Receipt{}, err
	}
	proposed, err := codec.DecodeAssignment(raw)
	if err != nil {
		return domain.Receipt{}, err
	}
	return a.facade.CreateAssignment(ctx, area, TeamFINC17(area).Issuer, proposed)
}

func (a *labApplication) SetGrantStatus(ctx context.Context, area domain.Area, fixtureContext domain.FixtureContext, proposed domain.GrantControl) (domain.GrantControl, error) {
	if area != a.area || fixtureContext.Name != labFixtureContext {
		return domain.GrantControl{}, domain.ErrRejected
	}
	if err := verifyMarker(ctx, a.path, area); err != nil {
		return domain.GrantControl{}, err
	}
	return a.facade.SetGrantStatus(ctx, area, TeamFINC17(area).Issuer, proposed)
}

func readOnlyDatabase(path string) (*sql.DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Join(domain.ErrUnavailable, err)
	}
	u := url.URL{Scheme: "file", Path: abs}
	query := u.Query()
	query.Set("mode", "ro")
	u.RawQuery = query.Encode()
	return sql.Open("sqlite", u.String())
}

func verifyMarker(ctx context.Context, path string, area domain.Area) error {
	db, err := readOnlyDatabase(path)
	if err != nil {
		return err
	}
	defer db.Close()
	var found int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_schema WHERE type='table' AND name='abv_lab_metadata'`).Scan(&found); err != nil {
		return errors.Join(domain.ErrUnavailable, err)
	}
	if found != 1 {
		return domain.ErrRejected
	}
	var version int
	var scenario, tenant, applicationID string
	err = db.QueryRowContext(ctx, `SELECT format_version,scenario_name,tenant_id,application_id FROM abv_lab_metadata WHERE marker='agentlabs-abv-lab'`).Scan(&version, &scenario, &tenant, &applicationID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrRejected
	}
	if err != nil {
		return errors.Join(domain.ErrUnavailable, err)
	}
	if version != 1 || scenario != labScenario {
		return domain.ErrUnsupported
	}
	if tenant != area.TenantID() || applicationID != area.ApplicationID() {
		return domain.ErrRejected
	}
	return nil
}
