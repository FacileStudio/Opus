// Package wireid is the id a request body carries: an int64 that also accepts
// the decimal string form, because both in-repo JavaScript clients normalise
// ids to strings before they reach a fetcher (see apps/web auth-client.ts toStr
// and packages/mcp tool args) while the OpenAPI documentation advertises int64.
// Numbers and numeric strings decode to the same value; anything else — empty
// string, float, overflow, garbage — is rejected and surfaces as the 400 the
// strict decoder already produced, so validation stays at the trust boundary.
package wireid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ID is an int64 id that unmarshals from a JSON number or a decimal string.
type ID int64

func (i *ID) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return fmt.Errorf("wireid: empty value")
	}
	if bytes.Equal(trimmed, []byte("null")) {
		*i = 0
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return fmt.Errorf("wireid: id must be a number or numeric string: %w", err)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return fmt.Errorf("wireid: id must not be empty")
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("wireid: %q is not an int64 id: %w", s, err)
		}
		*i = ID(v)
		return nil
	}
	v, err := strconv.ParseInt(string(trimmed), 10, 64)
	if err != nil {
		return fmt.Errorf("wireid: %s is not an int64 id: %w", trimmed, err)
	}
	*i = ID(v)
	return nil
}

func (i ID) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, int64(i), 10), nil
}

// Int64Ptr converts an optional wire id to the *int64 domain fields use,
// preserving nil for an absent key.
func Int64Ptr(p *ID) *int64 {
	if p == nil {
		return nil
	}
	v := int64(*p)
	return &v
}

// Int64s converts a wire id slice, for the bulk endpoints that bind ids into
// a query.
func Int64s(ids []ID) []int64 {
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		out = append(out, int64(id))
	}
	return out
}

// FromInt64s is the inverse of Int64s, for test and handler code holding a
// plain []int64.
func FromInt64s(ids []int64) []ID {
	out := make([]ID, 0, len(ids))
	for _, id := range ids {
		out = append(out, ID(id))
	}
	return out
}

// FromInt64Ptr is the inverse of Int64Ptr: a *int64 becomes a *ID, nil stays nil.
func FromInt64Ptr(p *int64) *ID {
	if p == nil {
		return nil
	}
	v := ID(*p)
	return &v
}
