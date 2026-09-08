package codec

import (
	"agentlabs.local/abv/domain"
	"encoding/json"
)

func DecodeAssignment(raw []byte) (domain.Assignment, error) {
	obj, err := object(raw)
	if err != nil {
		return domain.Assignment{}, err
	}
	if err = version(obj); err != nil {
		return domain.Assignment{}, err
	}
	if err = keys(obj, "version", "id", "grant_id", "grant_revision", "recipient", "status"); err != nil {
		return domain.Assignment{}, err
	}
	recipient, ok := obj["recipient"].(map[string]any)
	if !ok {
		return domain.Assignment{}, domain.ErrMalformed
	}
	if err = keys(recipient, "type", "id"); err != nil {
		return domain.Assignment{}, err
	}
	var result domain.Assignment
	if err = json.Unmarshal(raw, &result); err != nil {
		return domain.Assignment{}, domain.ErrMalformed
	}
	if invalidString(result.ID) || invalidString(result.GrantID) || result.GrantRevision <= 0 || invalidString(result.Recipient.ID) {
		return domain.Assignment{}, domain.ErrMalformed
	}
	if result.Status != "enabled" && result.Status != "disabled" {
		return domain.Assignment{}, domain.ErrMalformed
	}
	if result.Recipient.Type != "user" && result.Recipient.Type != "group" {
		return domain.Assignment{}, domain.ErrUnsupported
	}
	return result, nil
}
