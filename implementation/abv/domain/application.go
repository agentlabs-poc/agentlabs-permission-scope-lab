package domain

import (
	"strings"
	"unicode/utf8"
)

type Application struct{ id string }

func NewApplication(id string) (Application, error) {
	a := Application{id: id}
	if err := a.Validate(); err != nil {
		return Application{}, err
	}
	return a, nil
}
func (a Application) ID() string { return a.id }
func (a Application) Validate() error {
	if strings.TrimSpace(a.id) == "" || strings.Contains(a.id, "*") || !utf8.ValidString(a.id) {
		return ErrMalformed
	}
	return nil
}
