package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// CanonicalizeSpecJSON normalizes the JSON-valued parts of a validator spec so
// equivalent object representations have one stable compiled identity.
//
// JSON numbers are decoded with UseNumber to avoid float64 coercion. Arrays
// retain their order, while json.Marshal deterministically orders object keys
// and removes insignificant whitespace.
func CanonicalizeSpecJSON(spec Spec) (Spec, error) {
	spec = Normalize(spec)

	var err error
	spec.Options, err = canonicalizeRawJSON(spec.Options)
	if err != nil {
		return Spec{}, fmt.Errorf("canonicalize validator options: %w", err)
	}
	spec.Spec, err = canonicalizeRawJSON(spec.Spec)
	if err != nil {
		return Spec{}, fmt.Errorf("canonicalize validator spec: %w", err)
	}
	return spec, nil
}

func canonicalizeRawJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := requireJSONDecoderEOF(decoder); err != nil {
		return nil, err
	}

	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(canonical), nil
}

func requireJSONDecoderEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
