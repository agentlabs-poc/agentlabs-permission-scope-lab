// Package domain defines canonical core records and explicitly internal ABV values.
// Constructing or decoding a record does not establish usable authority.
package domain

import "strings"

// Area is the mandatory outer tenant/application boundary, never a scope predicate.
// Its zero value is invalid. Every authority entry point must call Validate.
type Area struct{ tenantID, applicationID string }

func NewArea(tenantID, applicationID string) (Area, error) {
	a := Area{tenantID: tenantID, applicationID: applicationID}
	if err := a.Validate(); err != nil {
		return Area{}, err
	}
	return a, nil
}
func (a Area) TenantID() string      { return a.tenantID }
func (a Area) ApplicationID() string { return a.applicationID }
func (a Area) Validate() error {
	for _, id := range []string{a.tenantID, a.applicationID} {
		if strings.TrimSpace(id) == "" || id == "*" {
			return ErrMalformed
		}
	}
	return nil
}
