package fielddsl

import "testing"

func TestLookupType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input      string
		wantGoType string
		wantSQL    string
	}{
		{input: "string", wantGoType: "string", wantSQL: "TEXT"},
		{input: "int", wantGoType: "int", wantSQL: "INTEGER"},
		{input: "int64", wantGoType: "int64", wantSQL: "BIGINT"},
		{input: "bool", wantGoType: "bool", wantSQL: "BOOLEAN"},
		{input: "float64", wantGoType: "float64", wantSQL: "DOUBLE PRECISION"},
		{input: "time", wantGoType: "time.Time", wantSQL: "TIMESTAMP WITH TIME ZONE"},
		{input: "string[]", wantGoType: "[]string", wantSQL: "TEXT[]"},
	}

	for _, tc := range testCases {
		spec, ok := LookupType(tc.input)
		if !ok {
			t.Fatalf("expected type %q to exist", tc.input)
		}
		if spec.GoType != tc.wantGoType {
			t.Fatalf("type %q: got GoType %q want %q", tc.input, spec.GoType, tc.wantGoType)
		}
		if spec.SQLType != tc.wantSQL {
			t.Fatalf("type %q: got SQLType %q want %q", tc.input, spec.SQLType, tc.wantSQL)
		}
	}
}

func TestLookupTypeUnsupported(t *testing.T) {
	t.Parallel()

	_, ok := LookupType("decimal")
	if ok {
		t.Fatal("expected decimal to be unsupported")
	}
}

func TestSupportedTypeKeysSorted(t *testing.T) {
	t.Parallel()

	keys := SupportedTypeKeys()
	if len(keys) == 0 {
		t.Fatal("expected non-empty supported type list")
	}

	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("expected keys to be sorted; %q came before %q", keys[i-1], keys[i])
		}
	}
}
