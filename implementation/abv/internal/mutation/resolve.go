package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"context"
	"time"
)

// validateReadingIdentity admits the actor types Q-086 approves — `user`,
// `agent` and `service_account` — where every write on this service still
// requires the actor to be the human.
//
// That difference is the point rather than a relaxation. An application
// enforcing its own endpoints asks about many humans and is none of them; a
// caller that had to *be* the subject could resolve only itself, which is enough
// to prove the read and useless to serve one. The Auth service already issues
// exactly this: a workload credential bound to one tenant application, acting as
// itself and naming the human it asks about.
//
// A `user` actor is still held to itself. A human acting is themselves, and a
// user actor naming a different human is impersonation rather than delegation —
// delegation arrives as an `agent` or `service_account`, whose right to ask is
// the gate's question.
func validateReadingIdentity(identity domain.Identity) error {
	if invalidIdentityPart(identity.Version) || invalidIdentityPart(identity.Actor.Type) ||
		invalidIdentityPart(identity.Actor.ID) || invalidIdentityPart(identity.HumanID) {
		return domain.ErrMalformed
	}
	if identity.Version != "1" {
		return domain.ErrUnsupported
	}
	switch identity.Actor.Type {
	case "user":
		if identity.Actor.ID != identity.HumanID {
			return domain.ErrRejected
		}
	case "agent", "service_account":
	default:
		return domain.ErrUnsupported
	}
	return nil
}

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
	if err := validateReadingIdentity(identity); err != nil {
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
