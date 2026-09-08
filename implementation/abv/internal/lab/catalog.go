package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"errors"
	"os"
	"time"
)

const catalogFixtureContext = "application-publisher"

var catalogPublisher = domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: catalogFixtureContext}, HumanID: catalogFixtureContext}

type catalogApplication struct {
	app    domain.Application
	path   string
	facade *abv.Facade
}

type catalogAdministration struct{ app domain.Application }

func (catalogAdministration) CheckAssignment(context.Context, abv.Evidence, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrRejected
}
func (a catalogAdministration) CheckPermissionRegistration(ctx context.Context, app domain.Application, catalog domain.Catalog, identity domain.Identity, _ domain.PermissionDefinition, _ []string, _ time.Time) error {
	return a.check(ctx, app, catalog, identity)
}
func (a catalogAdministration) CheckScopeRegistration(ctx context.Context, app domain.Application, catalog domain.Catalog, identity domain.Identity, _ domain.ScopeDefinition, _ time.Time) error {
	return a.check(ctx, app, catalog, identity)
}
func (a catalogAdministration) check(ctx context.Context, app domain.Application, catalog domain.Catalog, identity domain.Identity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if app != a.app || catalog.ApplicationID != a.app.ID() || identity != catalogPublisher {
		return domain.ErrRejected
	}
	return nil
}

var _ application.CatalogConnect = ConnectCatalog

func ConnectCatalog(ctx context.Context, app domain.Application, path string) (application.CatalogAPI, func() error, error) {
	if err := app.Validate(); err != nil {
		return nil, nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, nil, errors.Join(domain.ErrUnavailable, err)
	}
	facade, err := abv.OpenSQLite(ctx, path, catalogAdministration{app: app}, clock{})
	if err != nil {
		return nil, nil, err
	}
	return &catalogApplication{app: app, path: path, facade: facade}, facade.Close, nil
}

func (a *catalogApplication) RegisterPermission(ctx context.Context, app domain.Application, fixture domain.FixtureContext, definition domain.PermissionDefinition, keys []string) (domain.PermissionDefinition, error) {
	if app != a.app || fixture.Name != catalogFixtureContext {
		return domain.PermissionDefinition{}, domain.ErrRejected
	}
	if err := verifyCatalogMarker(ctx, a.path, app); err != nil {
		return domain.PermissionDefinition{}, err
	}
	return a.facade.RegisterPermission(ctx, app, catalogPublisher, definition, keys)
}
func (a *catalogApplication) RegisterScope(ctx context.Context, app domain.Application, fixture domain.FixtureContext, definition domain.ScopeDefinition) (domain.ScopeDefinition, error) {
	if app != a.app || fixture.Name != catalogFixtureContext {
		return domain.ScopeDefinition{}, domain.ErrRejected
	}
	if err := verifyCatalogMarker(ctx, a.path, app); err != nil {
		return domain.ScopeDefinition{}, err
	}
	return a.facade.RegisterScope(ctx, app, catalogPublisher, definition)
}

func verifyCatalogMarker(ctx context.Context, path string, app domain.Application) error {
	marker, err := loadMarker(ctx, path)
	if err != nil {
		return err
	}
	if marker.applicationID != app.ID() {
		return domain.ErrRejected
	}
	return nil
}
