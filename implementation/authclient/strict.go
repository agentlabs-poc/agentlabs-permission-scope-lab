package authclient

import (
	"bytes"
	"encoding/json"
	"errors"
)

// rejectAmbiguousJSON refuses an answer that does not mean exactly one thing.
//
// The service already refuses both of these in a question, and its reasoning
// applies unchanged in this direction: encoding/json takes the last occurrence
// of a repeated field silently, so an answer naming human_id twice is
// corroborated against the second while any log or proxy reading the first
// recorded the other — and a second document appended after the first is read
// by this client and not by whatever else is on the path.
//
// The client was strict about unknown fields and lax about both of these, which
// made the strictness decorative: a caller who can shape an answer could always
// say two things at once.
func rejectAmbiguousJSON(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := scanValue(decoder); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("trailing JSON")
	}
	return nil
}

func scanValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return errors.New("malformed object key")
			}
			if seen[name] {
				return errors.New("duplicate JSON field " + name)
			}
			seen[name] = true
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	}
	_, err = decoder.Token()
	return err
}
