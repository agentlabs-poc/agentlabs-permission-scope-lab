package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"context"
	"time"
)

// RootEstablishment gates bringing a root into existence, and it is deliberately
// not part of any grant administration interface.
//
// Q-113 requires that ordinary grant operations can never confer root authority.
// The guarantee is not a check inside grant creation — it is that no grant
// operation can reach this method. An administrator who may call every grant and
// assignment operation still cannot establish a root, because establishing one
// is not in their vocabulary.
//
// Which permission each gate demands is not chosen here. Q-113 leaves the
// procedure open: "The trusted operator/procedure, exact seed bounds ... proof
// of root establishment ... still need discussion." An adapter answers these,
// and the handbook can settle them later without changing the signatures.
type RootEstablishment interface {
	// CheckAuthRootEstablishment gates the tenant's Auth root. Its actor is Auth
	// platform administration: a tenant has no administrator until this runs, so
	// the authority cannot come from inside the tenant.
	CheckAuthRootEstablishment(context.Context, domain.Area, domain.Identity, string, time.Time) error
	// CheckRootEstablishment gates an application's root. Its actor is the
	// tenant administrator, holding Auth-boundary authority from the Auth root —
	// which is why it is not circular: the authority to create an application's
	// ceiling comes from outside that application entirely.
	CheckRootEstablishment(context.Context, domain.Area, domain.Identity, string, time.Time) error
}

func (s *Service) rootEstablishment(ctx context.Context, area domain.Area, identity domain.Identity) (RootEstablishment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(RootEstablishment)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}

// EstablishAuthRoot creates the tenant's authority over Auth itself — the first
// authority a tenant has, and the one that makes someone a tenant administrator.
//
// Its area names the platform's namespace rather than an application, so its
// computed ceiling is the platform catalog. There is no installation to require:
// Auth is not a registered application and is never installed.
func (s *Service) EstablishAuthRoot(ctx context.Context, area domain.Area, identity domain.Identity, holderTeamID string) (domain.Grant, domain.GrantContent, error) {
	return s.establish(ctx, area, identity, holderTeamID, true)
}

// EstablishRoot creates a tenant's ceiling inside one application, after the
// tenant has installed it.
func (s *Service) EstablishRoot(ctx context.Context, area domain.Area, identity domain.Identity, holderTeamID string) (domain.Grant, domain.GrantContent, error) {
	return s.establish(ctx, area, identity, holderTeamID, false)
}

func (s *Service) establish(ctx context.Context, area domain.Area, identity domain.Identity, holderTeamID string, auth bool) (domain.Grant, domain.GrantContent, error) {
	fail := func(err error) (domain.Grant, domain.GrantContent, error) {
		return domain.Grant{}, domain.GrantContent{}, err
	}
	admin, err := s.rootEstablishment(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	// The holder is a team, and team ids are base-36 Snowflakes — so a holder
	// that cannot be a team id is malformed rather than merely absent.
	if !codec.ValidRoleID(holderTeamID) {
		return fail(domain.ErrMalformed)
	}

	// Every identifier is issued, and a root's most of all: a caller who could
	// name a root is one step from claiming one.
	grant := domain.Grant{ID: s.ids.Next(), Status: "enabled", TrustedRoot: true}
	content := domain.GrantContent{
		Version: "1", GrantID: grant.ID, Revision: 1,
		// No parent and no permissions, both omitted rather than empty: the
		// trusted establishment supplies them, not this content. Scope {} is the
		// whole area, and a root narrower than its area is not a ceiling.
		Scope: map[string]string{},
	}
	assignment := domain.Assignment{
		Version: "1", ID: s.ids.Next(), GrantID: grant.ID, GrantRevision: 1,
		Recipient: domain.Recipient{Type: "group", ID: holderTeamID}, Status: "enabled",
	}

	err = s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		check := admin.CheckRootEstablishment
		if auth {
			check = admin.CheckAuthRootEstablishment
		}
		if err := check(ctx, area, identity, holderTeamID, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		established := storage.NewRoot{
			Grant: grant, Content: content, Assignment: assignment, HolderTeamID: holderTeamID,
		}
		return storage.WriteSet{NewRoot: &established}, nil
	})
	if err != nil {
		return fail(err)
	}
	return grant, content, nil
}
