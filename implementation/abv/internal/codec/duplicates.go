// Package codec decodes only the prototype's supported canonical core shapes.
// Decoding validates representation, not administrative permission or lineage.
package codec

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

// These are parser safety limits for this implementation, not canonical limits.
const maxInputBytes = 1 << 20
const maxDepth = 64

func object(raw []byte) (map[string]any, error) {
	if len(raw) > maxInputBytes {
		return nil, domain.ErrUnsupported
	}
	if !utf8.Valid(raw) || !validEscapedUnicode(raw) {
		return nil, domain.ErrMalformed
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	value, err := node(dec, 0)
	if err != nil {
		return nil, err
	}
	if _, err = dec.Token(); !errors.Is(err, io.EOF) {
		return nil, domain.ErrMalformed
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, domain.ErrMalformed
	}
	return obj, nil
}

// encoding/json replaces unpaired escaped UTF-16 surrogates with U+FFFD.
// Exact authority identifiers cannot permit that lossy repair.
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
			if i >= len(raw) {
				return true // the JSON decoder reports the truncated escape
			}
			if raw[i] != 'u' {
				continue
			}
			first, ok := hexCodeUnit(raw, i+1)
			if !ok {
				return true // the JSON decoder reports the malformed escape
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
func node(dec *json.Decoder, depth int) (any, error) {
	if depth > maxDepth {
		return nil, domain.ErrUnsupported
	}
	token, err := dec.Token()
	if err != nil || token == nil {
		return nil, domain.ErrMalformed
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delim {
	case '{':
		result := map[string]any{}
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return nil, domain.ErrMalformed
			}
			name, ok := key.(string)
			if !ok {
				return nil, domain.ErrMalformed
			}
			if _, exists := result[name]; exists {
				return nil, domain.ErrMalformed
			}
			value, err := node(dec, depth+1)
			if err != nil {
				return nil, err
			}
			result[name] = value
		}
		if end, err := dec.Token(); err != nil || end != json.Delim('}') {
			return nil, domain.ErrMalformed
		}
		return result, nil
	case '[':
		result := []any{}
		for dec.More() {
			value, err := node(dec, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		if end, err := dec.Token(); err != nil || end != json.Delim(']') {
			return nil, domain.ErrMalformed
		}
		return result, nil
	default:
		return nil, domain.ErrMalformed
	}
}
func keys(obj map[string]any, allowed ...string) error {
	for key := range obj {
		found := false
		for _, a := range allowed {
			if key == a {
				found = true
				break
			}
		}
		if !found {
			return domain.ErrUnsupported
		}
	}
	return nil
}
func version(obj map[string]any) error {
	v, ok := obj["version"].(string)
	if !ok || v == "" {
		return domain.ErrMalformed
	}
	if v != "1" {
		return domain.ErrUnsupported
	}
	return nil
}
