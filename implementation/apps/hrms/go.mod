module agentlabs.local/apps/hrms

go 1.25.1

require (
	agentlabs.local/authclient v0.0.0
	agentlabs.local/authmiddleware v0.0.0
)

replace agentlabs.local/authmiddleware => ../../authmiddleware

replace agentlabs.local/authclient => ../../authclient
