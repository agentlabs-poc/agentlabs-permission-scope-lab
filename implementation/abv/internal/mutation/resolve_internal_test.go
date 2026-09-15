package mutation

import (
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
)

// The identity rule, tested where it lives. Through the facade the lab gate
// refuses any user identity that is not the fixture administrator, so an
// impersonating identity is refused whether or not this rule exists — the
// assertion that looked like it covered this passed with the rule deleted.
//
// The rule matters because a deployment's gate will not be the fixture's: it may
// well admit many humans, and then this is the only thing standing between "ask
// about yourself" and "ask about anyone".
func TestValidateReadingIdentity(t *testing.T) {
	sound := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}

	for name, tc := range map[string]struct {
		identity domain.Identity
		want     error
	}{
		"a human asking about themselves": {sound, nil},
		"a credential asking about someone": {
			domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account", ID: "agent_hrms"}, HumanID: "fn2q6v8sbo1e"}, nil,
		},
		"an agent asking about someone": {
			domain.Identity{Version: "1", Actor: domain.Actor{Type: "agent", ID: "a-17"}, HumanID: "fn2q6v8sbo1e"}, nil,
		},
		// The one this file exists for.
		"a human naming somebody else": {
			domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fn2q6v8sbo1e"}, domain.ErrUnsupported,
		},
		"an actor type outside Q-086": {
			domain.Identity{Version: "1", Actor: domain.Actor{Type: "robot", ID: "r2"}, HumanID: "fi7io4lvjqio"}, domain.ErrUnsupported,
		},
		"a future version": {
			domain.Identity{Version: "2", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}, domain.ErrUnsupported,
		},
		"no subject":  {domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account", ID: "agent_hrms"}}, domain.ErrMalformed},
		"no actor id": {domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account"}, HumanID: "fi7io4lvjqio"}, domain.ErrMalformed},
		"wildcard subject": {
			domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account", ID: "agent_hrms"}, HumanID: "*"}, domain.ErrMalformed,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := validateReadingIdentity(tc.identity)
			if tc.want == nil && err != nil {
				t.Fatalf("refused a sound identity: %v", err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

// Writes are untouched: the actor must still be the human, whatever a read now
// admits. If these ever agree, a service credential can write authority.
func TestWritesStillRequireTheActorToBeTheHuman(t *testing.T) {
	credential := domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account", ID: "agent_hrms"}, HumanID: "fi7io4lvjqio"}
	if err := validateReadingIdentity(credential); err != nil {
		t.Fatalf("the read refused a credential: %v", err)
	}
	if err := validateSupportedIdentity(credential); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("a write admitted a credential: %v", err)
	}
}
