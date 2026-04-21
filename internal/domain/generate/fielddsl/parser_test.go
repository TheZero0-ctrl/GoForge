package fielddsl

import "testing"

func TestParseToken(t *testing.T) {
	t.Parallel()

	field, err := ParseToken("title:string")
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}

	if field.Name != "title" {
		t.Fatalf("unexpected field name: %q", field.Name)
	}

	if field.Type.Key != "string" {
		t.Fatalf("unexpected field type: %q", field.Type.Key)
	}
}

func TestParseTokenNormalizesCamelName(t *testing.T) {
	t.Parallel()

	field, err := ParseToken("publishedAt:time")
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}

	if field.Name != "published_at" {
		t.Fatalf("expected normalized snake_case name, got %q", field.Name)
	}
}

func TestParseTokenRejectsUnsupportedType(t *testing.T) {
	t.Parallel()

	_, err := ParseToken("price:decimal")
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
}

func TestParseManyRejectsDuplicates(t *testing.T) {
	t.Parallel()

	_, err := ParseMany([]string{"title:string", "title:int"})
	if err == nil {
		t.Fatal("expected duplicate field name error")
	}
}
