package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// teamBounds reads the shared paging flags. A malformed bound is malformed input
// rather than a silently applied default.
func teamBounds(flags map[string]string, offset, limit *int, diag io.Writer) int {
	for flag, target := range map[string]*int{"--offset": offset, "--limit": limit} {
		if has(flags, flag) {
			value, err := strconv.Atoi(flags[flag])
			if err != nil {
				return report(diag, domain.ErrMalformed)
			}
			*target = value
		}
	}
	return 0
}

// teamLines renders one team. The parent prints as an id, because that is what is
// stored — a name would be a second thing to keep in step.
func teamLines(team domain.Team) string {
	parent := team.ParentID
	if parent == "" {
		parent = "(root)"
	}
	return fmt.Sprintf("internal projection: team\nid  %s\nname  %s\nparent  %s\n", team.ID, team.Name, parent)
}

func renderTeam(out, diag io.Writer, team domain.Team) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	_, err := fmt.Fprint(out, teamLines(team))
	return err
}

func renderTeamPage(out, diag io.Writer, page domain.TeamPage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: teams\ncount  %d\ntotal  %d\ngeneration  %d\n",
		len(page.Teams), page.Total, page.Generation)
	for _, team := range page.Teams {
		parent := team.ParentID
		if parent == "" {
			parent = "(root)"
		}
		fmt.Fprintf(&rendered, "%s  %-18s parent=%s\n", team.ID, team.Name, parent)
	}
	_, err := out.Write(rendered.Bytes())
	return err
}

func renderMemberPage(out, diag io.Writer, page domain.MemberPage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: members\ncount  %d\ntotal  %d\ngeneration  %d\n",
		len(page.Members), page.Total, page.Generation)
	for _, m := range page.Members {
		fmt.Fprintf(&rendered, "team=%s  human=%s\n", m.TeamID, m.HumanID)
	}
	_, err := out.Write(rendered.Bytes())
	return err
}

func renderRemoved(out, diag io.Writer, kind, id string) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "internal projection: %s\nremoved  %s\n", kind, id)
	return err
}

func renderMembershipChange(out, diag io.Writer, verb, teamID, humanID string) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "internal projection: membership\n%s  team=%s  human=%s\n", verb, teamID, humanID)
	return err
}
