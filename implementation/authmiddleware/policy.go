package authmiddleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Source string

const (
	SourcePath Source = "path"
	SourceBody Source = "body"
)

type Input struct {
	Source Source `json:"source"`
	Name   string `json:"name"`
}

type Policy struct {
	Version    string           `json:"version"`
	Method     string           `json:"method"`
	Path       string           `json:"path"`
	Permission string           `json:"permission"`
	Inputs     map[string]Input `json:"inputs"`
}

func DecodePolicy(raw []byte) (Policy, error) {
	if err := validateJSON(raw); err != nil {
		return Policy{}, err
	}
	var policy Policy
	if err := strictDecode(raw, &policy); err != nil {
		return Policy{}, err
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (p Policy) Validate() error {
	if p.Version != "1" {
		return fmt.Errorf("unsupported policy version %q", p.Version)
	}
	if invalidText(p.Method) || p.Method != strings.ToUpper(p.Method) || !httpToken(p.Method) {
		return errors.New("invalid policy method")
	}
	placeholders, err := pathPlaceholders(p.Path)
	if err != nil {
		return err
	}
	if !validPermission(p.Permission) {
		return errors.New("invalid policy permission")
	}
	if p.Inputs == nil {
		return errors.New("missing policy inputs")
	}
	for local, input := range p.Inputs {
		if invalidText(local) || strings.ContainsAny(local, " \t\r\n") || invalidText(input.Name) {
			return errors.New("invalid policy input name")
		}
		switch input.Source {
		case SourcePath:
			if _, ok := placeholders[input.Name]; !ok {
				return fmt.Errorf("path input %q is not a declared placeholder", input.Name)
			}
		case SourceBody:
			if strings.ContainsAny(input.Name, ".[]/") {
				return fmt.Errorf("body input %q is not top-level", input.Name)
			}
		default:
			return fmt.Errorf("unsupported input source %q", input.Source)
		}
	}
	return nil
}

func validPermission(permission string) bool {
	parts := strings.Split(permission, "::")
	return !invalidText(permission) && len(parts) == 2 && parts[0] != "" && parts[1] != "" && !strings.ContainsAny(permission, "*, \t\r\n")
}

func pathPlaceholders(path string) (map[string]struct{}, error) {
	if invalidText(path) || path[0] != '/' || strings.ContainsAny(path, "?#") {
		return nil, errors.New("invalid absolute policy path")
	}
	result := map[string]struct{}{}
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '}':
			return nil, errors.New("invalid policy path placeholder")
		case '{':
			end := strings.IndexByte(path[i+1:], '}')
			if end < 0 {
				return nil, errors.New("invalid policy path placeholder")
			}
			end += i + 1
			name := path[i+1 : end]
			if invalidText(name) || strings.ContainsAny(name, "/{} \t\r\n") {
				return nil, errors.New("invalid policy path placeholder")
			}
			if _, exists := result[name]; exists {
				return nil, errors.New("duplicate policy path placeholder")
			}
			result[name] = struct{}{}
			i = end
		}
	}
	return result, nil
}

func invalidText(value string) bool {
	return value == "" || value != strings.TrimSpace(value) || !utf8.ValidString(value)
}

func httpToken(value string) bool {
	for _, r := range value {
		if !strings.ContainsRune("!#$%&'*+-.^_`|~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", r) {
			return false
		}
	}
	return true
}

const maxJSONBytes = 1 << 20
const maxJSONDepth = 64

func strictDecode(raw []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON")
	}
	return nil
}

func validateJSON(raw []byte) error {
	if len(raw) > maxJSONBytes {
		return errors.New("JSON exceeds 1 MiB")
	}
	if !utf8.Valid(raw) || !validEscapedUnicode(raw) {
		return errors.New("invalid JSON Unicode")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := jsonNode(decoder, 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON")
	}
	return nil
}

func jsonNode(decoder *json.Decoder, depth int) error {
	if depth > maxJSONDepth {
		return errors.New("JSON exceeds nesting limit")
	}
	token, err := decoder.Token()
	if err != nil || token == nil {
		return errors.New("invalid JSON")
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return errors.New("invalid JSON")
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		if delim == '{' {
			key, err := decoder.Token()
			name, ok := key.(string)
			if err != nil || !ok {
				return errors.New("invalid JSON object")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate JSON field %q", name)
			}
			seen[name] = struct{}{}
		}
		if err := jsonNode(decoder, depth+1); err != nil {
			return err
		}
	}
	end, err := decoder.Token()
	if err != nil || (delim == '{' && end != json.Delim('}')) || (delim == '[' && end != json.Delim(']')) {
		return errors.New("invalid JSON")
	}
	return nil
}

// encoding/json repairs unpaired escaped UTF-16 surrogates; reject them so
// identifiers and field names are preserved exactly.
func validEscapedUnicode(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '"' {
			continue
		}
		for i++; i < len(raw) && raw[i] != '"'; i++ {
			if raw[i] != '\\' {
				continue
			}
			i++
			if i >= len(raw) || raw[i] != 'u' {
				continue
			}
			first, ok := hexCodeUnit(raw, i+1)
			if !ok {
				continue
			}
			i += 4
			if first >= 0xdc00 && first <= 0xdfff {
				return false
			}
			if first < 0xd800 || first > 0xdbff {
				continue
			}
			if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
				return false
			}
			second, ok := hexCodeUnit(raw, i+3)
			if !ok || second < 0xdc00 || second > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}

func hexCodeUnit(raw []byte, start int) (uint16, bool) {
	if start+4 > len(raw) {
		return 0, false
	}
	var value uint16
	for _, b := range raw[start : start+4] {
		value <<= 4
		switch {
		case b >= '0' && b <= '9':
			value += uint16(b - '0')
		case b >= 'a' && b <= 'f':
			value += uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			value += uint16(b-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
