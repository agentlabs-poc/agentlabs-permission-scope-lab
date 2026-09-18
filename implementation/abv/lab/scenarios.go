package lab

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type Scenarios struct{}

var _ application.ScenarioRunner = Scenarios{}

func (Scenarios) Seed(ctx context.Context, area domain.Area, scenario, path string) error {
	switch scenario {
	case labScenario:
		// Two snapshots, because a tenant has two chains — Q-153's two namespaces
		// and Q-155's two roots. The business records are in the application's
		// area; the administrative records that authorize changing a team are in
		// the platform's, and no team write resolves without them.
		administrative, err := AuthAdministration(area)
		if err != nil {
			return err
		}
		return seedScenario(ctx, area, path, TeamFINC17(area).Snapshot, administrative)
	case genesisScenario:
		// Genesis is the moment before any authority exists, and that includes
		// administrative authority: there is no Auth chain to seed, which is why
		// establishing the Auth root is the first act a tenant can perform.
		return seedScenario(ctx, area, path, TenantGenesis(area))
	}
	return domain.ErrUnsupported
}

func (Scenarios) Run(ctx context.Context, area domain.Area, scenario, caseName, path string) error {
	if scenario != labScenario || caseName != "unsupported-permission" {
		return domain.ErrUnsupported
	}
	fixture := TeamFINC17(area)
	key := domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}
	unsupported := fixture.Snapshot.Contents[key]
	unsupported.Permissions = []string{PayslipDelete}
	unsupported.Scope = map[string]string{}
	fixture.Snapshot.Contents[key] = unsupported
	administrative, err := AuthAdministration(area)
	if err != nil {
		return err
	}
	if err := seedScenario(ctx, area, path, fixture.Snapshot, administrative); err != nil {
		return err
	}
	api, closeConnection, err := Connect(ctx, area, path)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(fixture.Proposed)
	if err == nil {
		_, err = api.Assign(ctx, area, domain.FixtureContext{Name: labFixtureContext}, raw)
	}
	if !errors.Is(err, domain.ErrRejected) {
		_ = closeConnection()
		if err == nil {
			return errors.New("unsupported permission assignment succeeded")
		}
		return err
	}
	_, inspectErr := api.Inspect(ctx, area, "assignment", fixture.Proposed.ID)
	closeErr := closeConnection()
	if !errors.Is(inspectErr, domain.ErrNotFound) {
		return errors.New("rejected scenario wrote assignment")
	}
	return closeErr
}

func seedScenario(ctx context.Context, area domain.Area, path string, snapshots ...storage.Snapshot) error {
	provider, err := CreateSQLite(ctx, path, snapshots)
	if err != nil {
		return err
	}
	if err = provider.Close(); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		return errors.Join(domain.ErrUnavailable, err)
	}
	defer db.Close()
	_, err = db.ExecContext(ctx, `
BEGIN IMMEDIATE;
CREATE TABLE abv_lab_metadata (
  marker TEXT PRIMARY KEY CHECK(marker='agentlabs-abv-lab'),
  format_version INTEGER NOT NULL,
  scenario_name TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  application_id TEXT NOT NULL
);
INSERT INTO abv_lab_metadata(marker,format_version,scenario_name,tenant_id,application_id)
VALUES('agentlabs-abv-lab',1,?,?,?);
COMMIT;`, labScenario, area.TenantID(), area.ApplicationID())
	if err != nil {
		return errors.Join(domain.ErrUnavailable, err)
	}
	return nil
}
