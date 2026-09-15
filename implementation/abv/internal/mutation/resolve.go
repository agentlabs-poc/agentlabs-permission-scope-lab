package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"context"
	"time"
)

// AuthorityRead gates reading what a human is entitled to.
//
// It is its own gate rather than a reuse of CheckGrantRead, because the question
// is different: not "may you see this grant record" but "may you ask what this
// human holds here". An application enforcing its own endpoints must be able to
// ask; a tenant administrator browsing records need not be the same answer.
type AuthorityRead interface {
	CheckAuthorityRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

// ResolveAuthority answers what one human is entitled to inside one area.
//
// It is the enforcement side's only read, and the first non-administrative one
// on this service. Everything else here changes authority; this one uses it.
//
// The identity block names both parties: Actor is the caller and HumanID is the
// subject whose authority is resolved. Q-086 admits `user`, `agent` and
// `service_account` actors, and this signature is already the shape that carries
// them — but today they must agree, because validateSupportedIdentity refuses an
// actor that differs from the subject, as every operation here does.
//
// That is the one thing an enforcing client will need lifted: an application
// asks about many humans and is not any of them. Until it is, a caller can only
// resolve its own authority, which is enough to prove the read and not enough to
// serve an application. Lifting it is delegation work (AUTHORITY-002), not a
// change to what this returns.
func (s *Service) ResolveAuthority(ctx context.Context, area domain.Area, identity domain.Identity, opts domain.ResolveOptions) (domain.ResolvedAuthority, error) {
	fail := func(err error) (domain.ResolvedAuthority, error) {
		return domain.ResolvedAuthority{}, err
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	admin, ok := s.administration.(AuthorityRead)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	var resolved domain.ResolvedAuthority
	err := s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckAuthorityRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		answer, err := lineage.ResolveAuthority(ctx, snapshot, identity, opts, s.clock.Now())
		if err != nil {
			return err
		}
		resolved = answer
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return resolved, nil
}
