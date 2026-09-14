package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// grantLines renders a head. The root marker is printed because it is the one
// thing about a grant a reader cannot infer from anything else on the row.
func grantLines(grant domain.Grant) string {
	kind := "child"
	if grant.TrustedRoot {
		kind = "root"
	}
	return fmt.Sprintf("internal projection: grant\nid  %s\nstatus  %s\nkind  %s\n", grant.ID, grant.Status, kind)
}

// contentLines renders a revision. A root prints "(computed from the catalog)"
// where a child prints its selection, because that is exactly the difference:
// the root stores no list, and printing an empty one would read as "none".
func contentLines(content domain.GrantContent) string {
	var b strings.Builder
	fmt.Fprintf(&b, "revision  %d\n", content.Revision)
	if content.ParentGrantID == "" {
		fmt.Fprint(&b, "parent  (none — root)\n")
	} else {
		fmt.Fprintf(&b, "parent  %s\n", content.ParentGrantID)
	}
	switch {
	case content.Permissions != nil:
		fmt.Fprintf(&b, "permissions  %s\n", strings.Join(content.Permissions, ", "))
	case content.RoleID != "":
		fmt.Fprintf(&b, "role  %s revision %d\n", content.RoleID, content.RoleRevision)
	default:
		fmt.Fprint(&b, "permissions  (computed from the catalog)\n")
	}
	keys := make([]string, 0, len(content.Scope))
	for key := range content.Scope {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		fmt.Fprint(&b, "scope  {} (adds no narrowing)\n")
		return b.String()
	}
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+content.Scope[key])
	}
	fmt.Fprintf(&b, "scope  %s\n", strings.Join(pairs, " AND "))
	return b.String()
}

func renderGrant(out, diag io.Writer, grant domain.Grant, content domain.GrantContent) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	rendered.WriteString(grantLines(grant))
	if content.GrantID != "" {
		rendered.WriteString(contentLines(content))
	}
	_, err := out.Write(rendered.Bytes())
	return err
}

func renderGrantPage(out, diag io.Writer, page domain.GrantPage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: grants\ncount  %d\ntotal  %d\n", len(page.Grants), page.Total)
	for _, grant := range page.Grants {
		kind := "child"
		if grant.TrustedRoot {
			kind = "root"
		}
		fmt.Fprintf(&rendered, "%-14s %-9s %s\n", grant.ID, grant.Status, kind)
	}
	_, err := out.Write(rendered.Bytes())
	return err
}

func renderRevisionPage(out, diag io.Writer, id string, page domain.GrantRevisionPage) error {
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
		return err
	}
	var rendered bytes.Buffer
	fmt.Fprintf(&rendered, "internal projection: grant revisions\ngrant  %s\ncount  %d\ntotal  %d\n",
		id, len(page.Revisions), page.Total)
	for _, content := range page.Revisions {
		source := "(computed)"
		switch {
		case content.Permissions != nil:
			source = strconv.Itoa(len(content.Permissions)) + " permissions"
		case content.RoleID != "":
			source = "role " + content.RoleID
		}
		fmt.Fprintf(&rendered, "revision %-4d %s\n", content.Revision, source)
	}
	_, err := out.Write(rendered.Bytes())
	return err
}

// grantScope parses repeated --scope key=value pairs. A malformed pair is
// malformed input rather than a silently dropped constraint.
func grantScope(raw string) (map[string]string, error) {
	scope := map[string]string{}
	if raw == "" {
		return scope, nil
	}
	for _, pair := range strings.Split(raw, ",") {
		key, value, found := strings.Cut(pair, "=")
		if !found || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return nil, domain.ErrMalformed
		}
		scope[key] = value
	}
	return scope, nil
}
