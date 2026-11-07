package wireid

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalAcceptsNumberAndNumericString(t *testing.T) {
	cases := []struct {
		in   string
		want ID
	}{
		{`7`, 7},
		{`"7"`, 7},
		{`" 7 "`, 7},
		{`-12`, -12},
		{`"9223372036854775807"`, 9223372036854775807},
		{`null`, 0},
	}
	for _, c := range cases {
		var got ID
		if err := json.Unmarshal([]byte(c.in), &got); err != nil {
			t.Fatalf("Unmarshal(%s): unexpected error %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("Unmarshal(%s) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestUnmarshalRejectsInvalid(t *testing.T) {
	for _, in := range []string{`""`, `"abc"`, `1.5`, `1e3`, `true`, `9223372036854775808`, `{}`, `[1]`} {
		var got ID
		if err := json.Unmarshal([]byte(in), &got); err == nil {
			t.Fatalf("Unmarshal(%s) = %d, want error", in, got)
		}
	}
}

func TestPointerAndSliceHelpers(t *testing.T) {
	var absent *ID
	if Int64Ptr(absent) != nil {
		t.Fatalf("Int64Ptr(nil) = %v, want nil", Int64Ptr(absent))
	}
	present := ID(9)
	if v := Int64Ptr(&present); v == nil || *v != 9 {
		t.Fatalf("Int64Ptr(&9) = %v, want pointer to 9", Int64Ptr(&present))
	}
	if got := Int64s([]ID{1, 2}); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("Int64s = %v, want [1 2]", got)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	data, err := json.Marshal(ID(42))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != "42" {
		t.Fatalf("Marshal(ID(42)) = %s, want 42 (bare number on the wire)", data)
	}
}
