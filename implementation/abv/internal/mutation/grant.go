package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"context"
	"sort"
	"time"
)

const (
	defaultGrantPage = 100
	maxGrantPage     = 500
)

// GrantAdministration gates creating and destroying a grant. They are separate
// authorities from publishing a revision, which has its own check: amending a
// grant a tenant already holds is not the same act as bringing a new one into
// existence, or ending one.
//
// Establishing a *root* is deliberately absent. Q-113 requires that ordinary
// grant administration can never confer root authority, and the way to guarantee
// that is not a check inside creation — it is that no grant operation writes
// trust evidence at all.
type GrantAdministration interface {
	CheckGrantCreate(context.Context, domain.Area, domain.Identity, domain.GrantContent, time.Time) error
	CheckGrantDelete(context.Context, domain.Area, domain.Identity, string, time.Time) error
}

// GrantReadAdministration gates reads of an area's grants.
type GrantReadAdministration interface {
	CheckGrantRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

func (s *Service) grantAdministration(ctx context.Context, area domain.Area, identity domain.Identity) (GrantAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(GrantAdministration)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}

func (s *Service) grantCatalog(ctx context.Context, area domain.Area, identity domain.Identity) (GrantReadAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(GrantReadAdministration)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}

// CreateGrant brings a grant into existence: its head and its first revision,
// written together.
//
// It is one operation rather than a create followed by a publish because
// revision 1 cannot go through the publish path — that path amends a grant and
// requires a predecessor. A head with no content would be a grant that can be
// enabled and supplies nothing.
//
// The id is issued, never accepted. A caller who could name a grant could name
// one that already exists somewhere it should not, and for a root it would be
// one step from claiming one.
func (s *Service) CreateGrant(ctx context.Context, area domain.Area, identity domain.Identity, parentGrantID string, proposed domain.GrantContent) (domain.Grant, domain.GrantContent, error) {
	fail := func(err error) (domain.Grant, domain.GrantContent, error) {
		return domain.Grant{}, domain.GrantContent{}, err
	}
	admin, err := s.grantAdministration(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if parentGrantID == "" {
		return fail(domain.ErrRejected)
	}
	grant := domain.Grant{ID: s.ids.Next(), Status: "enabled"}
	content := proposed
	content.Version, content.GrantID, content.Revision = "1", grant.ID, 1
	content.ParentGrantID = parentGrantID
	// No missing-scope default. The handbook forbids it by name — "omitting
	// scope or supplying null remains invalid; no missing-scope default to {}
	// is permitted" — and codec.ValidateContent one line below already refuses a
	// nil scope, which is how PublishGrantRevision answers the same input. This
	// substitution stood between the two, so a caller that failed to populate
	// scope was given the widest child its parent permits instead of a refusal.
	if err := codec.ValidateContent(content); err != nil {
		return fail(err)
	}
	err = s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckGrantCreate(ctx, area, identity, content, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		created := storage.NewGrant{Grant: grant, Content: content}
		return storage.WriteSet{NewGrant: &created}, nil
	})
	if err != nil {
		return fail(err)
	}
	return grant, content, nil
}

// DeleteGrant destroys a grant whole — head and every revision. Q-082 folded
// revoke into delete, so this is the one permanent removal.
//
// It refuses a grant that anything depends on, the rule teams settled. Until
// assignments are their own record the dependants it can see are children, and
// the refusal is deliberately the conservative direction: it is loosened when
// assignments can be consulted, never tightened.
func (s *Service) DeleteGrant(ctx context.Context, area domain.Area, identity domain.Identity, id string) error {
	admin, err := s.grantAdministration(ctx, area, identity)
	if err != nil {
		return err
	}
	if id == "" {
		return domain.ErrMalformed
	}
	return s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckGrantDelete(ctx, area, identity, id, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if _, ok := snapshot.Controls[id]; !ok {
			return storage.WriteSet{}, domain.ErrNotFound
		}
		// A lab default, not an agreed rule, and the handbook leans the other
		// way on the neighbouring question.
		//
		// Q-113 (bootstrap-authority.md:134-136) is about *conferring* root
		// authority: "ordinary grant creation or modification must not confer
		// root authority merely by omitting/removing a parent reference." It
		// says nothing about removing a root that was properly established, and
		// bootstrap-authority.md:151-152 says the opposite of special: an
		// established root "remains an ordinary grant subject to status,
		// validity, revisions, assignments, and its explicit boundaries."
		// Deletion is not in that list, which is the only reason this refusal is
		// not flatly against the text — and the lab already refuses to *disable*
		// a root, which is.
		//
		// It stays because an area whose root is gone has no ceiling for
		// anything, recovery is an unbuilt contract, and nothing in this lab
		// deletes a root — so the cost of being wrong is an operation nobody
		// performs. The authorized root-change procedure is open
		// (root-grant-format.md:90-91) and the service this migrates into is
		// where it gets written. Recorded in plan/migration-requirements.md.
		if snapshot.TrustedRoots[id] {
			return storage.WriteSet{}, domain.ErrUnsupported
		}
		for _, content := range snapshot.Contents {
			if content.ParentGrantID == id {
				return storage.WriteSet{}, domain.ErrConflict
			}
		}
		for _, assignment := range snapshot.Assignments {
			if assignment.GrantID == id {
				return storage.WriteSet{}, domain.ErrConflict
			}
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		return storage.WriteSet{RemovedGrant: id}, nil
	})
}

// GetGrant returns the head, and one revision when asked for. revision 0 means
// the head alone: a caller checking whether a grant is enabled should not have
// to name a revision to find out.
func (s *Service) GetGrant(ctx context.Context, area domain.Area, identity domain.Identity, id string, revision int64) (domain.Grant, domain.GrantContent, error) {
	fail := func(err error) (domain.Grant, domain.GrantContent, error) {
		return domain.Grant{}, domain.GrantContent{}, err
	}
	admin, err := s.grantCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if id == "" || revision < 0 {
		return fail(domain.ErrMalformed)
	}
	var grant domain.Grant
	var content domain.GrantContent
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckGrantRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		control, ok := snapshot.Controls[id]
		if !ok || control.ID != id {
			return domain.ErrNotFound
		}
		grant = domain.Grant{ID: id, Status: control.Status, TrustedRoot: snapshot.TrustedRoots[id]}
		if revision == 0 {
			return nil
		}
		found, ok := snapshot.Contents[domain.GrantKey{ID: id, Revision: revision}]
		if !ok {
			return domain.ErrNotFound
		}
		content = found
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return grant, content, nil
}

func (s *Service) ListGrants(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.GrantFilter) (domain.GrantPage, error) {
	fail := func(err error) (domain.GrantPage, error) { return domain.GrantPage{}, err }
	admin, err := s.grantCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxGrantPage {
		return fail(domain.ErrMalformed)
	}
	if filter.Status != "" && filter.Status != "enabled" && filter.Status != "disabled" {
		return fail(domain.ErrMalformed)
	}
	var page domain.GrantPage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckGrantRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.Grant, 0, len(snapshot.Controls))
		for id, control := range snapshot.Controls {
			if err := ctx.Err(); err != nil {
				return err
			}
			if control.ID != id {
				return domain.ErrRejected
			}
			root := snapshot.TrustedRoots[id]
			if filter.Status != "" && control.Status != filter.Status {
				continue
			}
			if filter.Root != nil && *filter.Root != root {
				continue
			}
			matched = append(matched, domain.Grant{ID: id, Status: control.Status, TrustedRoot: root})
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })
		page = domain.GrantPage{Grants: grantPage(matched, filter.Offset, filter.Limit), Total: len(matched)}
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

// ListGrantRevisions returns one grant's revisions, newest first — the order a
// reader wants, because the latest is what a new assignment would adopt.
func (s *Service) ListGrantRevisions(ctx context.Context, area domain.Area, identity domain.Identity, id string, offset, limit int) (domain.GrantRevisionPage, error) {
	fail := func(err error) (domain.GrantRevisionPage, error) { return domain.GrantRevisionPage{}, err }
	admin, err := s.grantCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if id == "" || offset < 0 || limit < 0 || limit > maxGrantPage {
		return fail(domain.ErrMalformed)
	}
	var page domain.GrantRevisionPage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckGrantRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		if _, ok := snapshot.Controls[id]; !ok {
			return domain.ErrNotFound
		}
		matched := make([]domain.GrantContent, 0, 4)
		for key, content := range snapshot.Contents {
			if err := ctx.Err(); err != nil {
				return err
			}
			if key.ID == id {
				matched = append(matched, content)
			}
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Revision > matched[j].Revision })
		page = domain.GrantRevisionPage{Revisions: grantPage(matched, offset, limit), Total: len(matched)}
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

func grantPage[T any](items []T, offset, limit int) []T {
	if limit == 0 {
		limit = defaultGrantPage
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
