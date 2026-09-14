// Package registry is the application registry: which applications exist, and
// which tenants have them installed.
//
// It is a 123 domain of its own. It does not import Auth-AL, does not know what
// a permission or a grant is, and holds no opinion about what an application's
// records mean — only that the application exists and who has it.
package registry

import (
	"agentlabs.local/registry/domain"
	"agentlabs.local/registry/internal/storage/sqlite"
	"context"
	"time"
)

// Administration is the injected gate. The registry validates shape and
// consistency; whether this human may register an application or install one for
// a tenant is the platform's decision, not this domain's.
//
// The three methods are three authorities, deliberately: registering an
// application is the platform acting, installing one is a tenant administrator
// acting, and collapsing them would make one act authorise the other.
type Administration interface {
	CheckApplicationWrite(ctx context.Context, identity domain.Identity, slug string, now time.Time) error
	CheckApplicationRead(ctx context.Context, identity domain.Identity, now time.Time) error
	CheckInstallationWrite(ctx context.Context, identity domain.Identity, tenantID, slug string, now time.Time) error
}

// Clock is a seam so a test does not depend on the wall clock.
type Clock interface{ Now() time.Time }

type Facade struct {
	store *sqlite.Store
	admin Administration
	clock Clock
}

const (
	defaultPage = 100
	maxPage     = 500
)

func Open(ctx context.Context, path string, admin Administration, clock Clock, create bool) (*Facade, error) {
	if admin == nil || clock == nil {
		return nil, domain.ErrMalformed
	}
	store, err := sqlite.Open(ctx, path, create)
	if err != nil {
		return nil, err
	}
	return &Facade{store: store, admin: admin, clock: clock}, nil
}

func (f *Facade) Close() error { return f.store.Close() }

// RegisterApplication is the only way an application comes into existence, and
// it is add-only: a slug is never reassigned, because it appears in the
// canonical path of every record that application owns in every domain, and
// reusing one would silently re-point them.
func (f *Facade) RegisterApplication(ctx context.Context, identity domain.Identity, slug, name string) (domain.Application, error) {
	fail := func(err error) (domain.Application, error) { return domain.Application{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if !identity.Valid() {
		return fail(domain.ErrUnsupported)
	}
	if !domain.ValidSlug(slug) || !domain.ValidName(name) {
		return fail(domain.ErrMalformed)
	}
	if err := f.admin.CheckApplicationWrite(ctx, identity, slug, f.clock.Now()); err != nil {
		return fail(err)
	}
	app := domain.Application{Slug: slug, Name: name, Status: domain.StatusActive}
	if err := f.store.InsertApplication(ctx, app); err != nil {
		return fail(err)
	}
	return app, nil
}

// SetApplicationStatus suspends or reactivates. It is reversible and never
// creates: a status change must not be a back door around registration.
func (f *Facade) SetApplicationStatus(ctx context.Context, identity domain.Identity, slug, status string) (domain.Application, error) {
	fail := func(err error) (domain.Application, error) { return domain.Application{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if !identity.Valid() {
		return fail(domain.ErrUnsupported)
	}
	if !domain.ValidSlug(slug) || !domain.ValidStatus(status) {
		return fail(domain.ErrMalformed)
	}
	if err := f.admin.CheckApplicationWrite(ctx, identity, slug, f.clock.Now()); err != nil {
		return fail(err)
	}
	return f.store.UpdateApplicationStatus(ctx, slug, status)
}

func (f *Facade) GetApplication(ctx context.Context, identity domain.Identity, slug string) (domain.Application, error) {
	fail := func(err error) (domain.Application, error) { return domain.Application{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if !identity.Valid() {
		return fail(domain.ErrUnsupported)
	}
	if !domain.ValidSlug(slug) {
		return fail(domain.ErrMalformed)
	}
	if err := f.admin.CheckApplicationRead(ctx, identity, f.clock.Now()); err != nil {
		return fail(err)
	}
	return f.store.Application(ctx, slug)
}

func (f *Facade) ListApplications(ctx context.Context, identity domain.Identity, filter domain.ApplicationFilter) (domain.ApplicationPage, error) {
	fail := func(err error) (domain.ApplicationPage, error) { return domain.ApplicationPage{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if !identity.Valid() {
		return fail(domain.ErrUnsupported)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxPage {
		return fail(domain.ErrMalformed)
	}
	if filter.Status != "" && !domain.ValidStatus(filter.Status) {
		return fail(domain.ErrMalformed)
	}
	if err := f.admin.CheckApplicationRead(ctx, identity, f.clock.Now()); err != nil {
		return fail(err)
	}
	all, err := f.store.Applications(ctx)
	if err != nil {
		return fail(err)
	}
	matched := all[:0:0]
	for _, app := range all {
		if filter.Status != "" && app.Status != filter.Status {
			continue
		}
		matched = append(matched, app)
	}
	return domain.ApplicationPage{Applications: page(matched, filter.Offset, filter.Limit), Total: len(matched)}, nil
}

// Install records that a tenant holds an application. It is add-only and
// requires the application to exist — the foreign key a single schema gave for
// free becomes a rule once the domains are separate.
func (f *Facade) Install(ctx context.Context, identity domain.Identity, tenantID, slug string) error {
	if err := f.installationPrecondition(ctx, identity, tenantID, slug); err != nil {
		return err
	}
	if _, err := f.store.Application(ctx, slug); err != nil {
		return err
	}
	// An installation is created enabled. Installing over one that exists —
	// enabled or disabled — is a conflict rather than a quiet re-enable: install
	// means "create the relationship", and reviving a disabled one is
	// SetInstallationStatus. One operation, one meaning.
	return f.store.InsertInstallation(ctx, domain.Installation{TenantID: tenantID, Slug: slug, Status: domain.StatusEnabled})
}

// Uninstall removes one tenant's hold on one application, and says whether it
// did anything.
//
// It does NOT refuse while the tenant holds live authority: the dependants live
// in Auth-AL's domain, which this one cannot see. Auth-AL stops resolving when
// the installation goes, which is the behaviour it already has. Guarding it
// would need a question asked back across the seam, and that is deliberately
// not decided here.
func (f *Facade) Uninstall(ctx context.Context, identity domain.Identity, tenantID, slug string) error {
	if err := f.installationPrecondition(ctx, identity, tenantID, slug); err != nil {
		return err
	}
	return f.store.DeleteInstallation(ctx, domain.Installation{TenantID: tenantID, Slug: slug})
}

// SetInstallationStatus enables or disables one tenant's installation.
//
// It is reversible and it never creates. Disabling is not uninstalling: the
// record survives, which is what lets a tenant pause an application without
// losing it. Gated as a tenant write, because disabling your own installation is
// the tenant administrator acting — the same authority as installing.
func (f *Facade) SetInstallationStatus(ctx context.Context, identity domain.Identity, tenantID, slug, status string) (domain.Installation, error) {
	fail := func(err error) (domain.Installation, error) { return domain.Installation{}, err }
	if !domain.ValidInstallationStatus(status) {
		return fail(domain.ErrMalformed)
	}
	if err := f.installationPrecondition(ctx, identity, tenantID, slug); err != nil {
		return fail(err)
	}
	return f.store.UpdateInstallationStatus(ctx, tenantID, slug, status)
}

// GetInstallation returns one tenant's hold on one application, with its status.
func (f *Facade) GetInstallation(ctx context.Context, identity domain.Identity, tenantID, slug string) (domain.Installation, error) {
	fail := func(err error) (domain.Installation, error) { return domain.Installation{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if !identity.Valid() {
		return fail(domain.ErrUnsupported)
	}
	if !domain.ValidTenant(tenantID) || !domain.ValidSlug(slug) {
		return fail(domain.ErrMalformed)
	}
	if err := f.admin.CheckApplicationRead(ctx, identity, f.clock.Now()); err != nil {
		return fail(err)
	}
	return f.store.Installation(ctx, tenantID, slug)
}

// IsInstalled reports whether the tenant holds the application AND has it
// enabled. Two facts, one bit — the caller's question is whether the tenant may
// use it, and a disabled installation answers no as surely as an absent one.
func (f *Facade) IsInstalled(ctx context.Context, identity domain.Identity, tenantID, slug string) (bool, error) {
	installation, err := f.GetInstallation(ctx, identity, tenantID, slug)
	if errorIs(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return installation.Status == domain.StatusEnabled, nil
}

// ListInstallations answers in both directions and requires exactly one filter.
// An unfiltered listing of every installation is unbounded in the dimension that
// grows fastest, and is not a question anyone asks.
func (f *Facade) ListInstallations(ctx context.Context, identity domain.Identity, filter domain.InstallationFilter) (domain.InstallationPage, error) {
	fail := func(err error) (domain.InstallationPage, error) { return domain.InstallationPage{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if !identity.Valid() {
		return fail(domain.ErrUnsupported)
	}
	if (filter.TenantID == "") == (filter.Slug == "") {
		return fail(domain.ErrMalformed)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxPage {
		return fail(domain.ErrMalformed)
	}
	if filter.TenantID != "" && !domain.ValidTenant(filter.TenantID) {
		return fail(domain.ErrMalformed)
	}
	if filter.Slug != "" && !domain.ValidSlug(filter.Slug) {
		return fail(domain.ErrMalformed)
	}
	if err := f.admin.CheckApplicationRead(ctx, identity, f.clock.Now()); err != nil {
		return fail(err)
	}
	all, err := f.store.Installations(ctx, filter.TenantID, filter.Slug)
	if err != nil {
		return fail(err)
	}
	return domain.InstallationPage{Installations: page(all, filter.Offset, filter.Limit), Total: len(all)}, nil
}

func (f *Facade) installationPrecondition(ctx context.Context, identity domain.Identity, tenantID, slug string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !identity.Valid() {
		return domain.ErrUnsupported
	}
	if !domain.ValidTenant(tenantID) || !domain.ValidSlug(slug) {
		return domain.ErrMalformed
	}
	return f.admin.CheckInstallationWrite(ctx, identity, tenantID, slug, f.clock.Now())
}

// page bounds a listing. An offset past the end is an empty page, never an
// error: a reader that jumps beyond the last page sees nothing rather than
// failing.
func page[T any](items []T, offset, limit int) []T {
	if limit == 0 {
		limit = defaultPage
	}
	if offset >= len(items) {
		return []T{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
