package main

import (
	regdomain "agentlabs.local/registry/domain"
	"context"
	"time"
)

func operatorIdentity() regdomain.Identity {
	return regdomain.Identity{Version: "1", HumanID: "fi7io4lvjqio"}
}

// labRegistryAdmin admits the lab's operator. A deployment gates these with its
// own platform administration.
type labRegistryAdmin struct{}

func (labRegistryAdmin) CheckApplicationWrite(_ context.Context, i regdomain.Identity, _ string, _ time.Time) error {
	return allow(i)
}
func (labRegistryAdmin) CheckApplicationRead(_ context.Context, i regdomain.Identity, _ time.Time) error {
	return allow(i)
}
func (labRegistryAdmin) CheckInstallationWrite(_ context.Context, i regdomain.Identity, _, _ string, _ time.Time) error {
	return allow(i)
}

func allow(i regdomain.Identity) error {
	if i != operatorIdentity() {
		return regdomain.ErrRejected
	}
	return nil
}
