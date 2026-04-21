package naming

import "testing"

func TestToSnake(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  string
	}{
		{input: "movie", want: "movie"},
		{input: "Movie", want: "movie"},
		{input: "MovieReview", want: "movie_review"},
		{input: "movie-review", want: "movie_review"},
		{input: "movie review", want: "movie_review"},
		{input: "HTTPServer", want: "http_server"},
	}

	for _, tc := range testCases {
		if got := ToSnake(tc.input); got != tc.want {
			t.Fatalf("ToSnake(%q): got %q want %q", tc.input, got, tc.want)
		}
	}
}

func TestPluralizeSingularize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		singular string
		plural   string
	}{
		{singular: "movie", plural: "movies"},
		{singular: "box", plural: "boxes"},
		{singular: "category", plural: "categories"},
		{singular: "person", plural: "people"},
	}

	for _, tc := range testCases {
		if got := Pluralize(tc.singular); got != tc.plural {
			t.Fatalf("Pluralize(%q): got %q want %q", tc.singular, got, tc.plural)
		}
		if got := Singularize(tc.plural); got != tc.singular {
			t.Fatalf("Singularize(%q): got %q want %q", tc.plural, got, tc.singular)
		}
	}
}

func TestNormalizeResourceName(t *testing.T) {
	t.Parallel()

	names, err := NormalizeResourceName("movies")
	if err != nil {
		t.Fatalf("NormalizeResourceName: %v", err)
	}

	if names.Singular != "movie" || names.Plural != "movies" {
		t.Fatalf("unexpected singular/plural: %+v", names)
	}

	if names.SingularPascal != "Movie" || names.PluralPascal != "Movies" {
		t.Fatalf("unexpected pascal names: %+v", names)
	}
}

func TestIsSnakeIdentifier(t *testing.T) {
	t.Parallel()

	valid := []string{"title", "published_at", "v1_name", "name2"}
	for _, value := range valid {
		if !IsSnakeIdentifier(value) {
			t.Fatalf("expected valid snake identifier: %q", value)
		}
	}

	invalid := []string{"", "Title", "published-at", "9name", "name!"}
	for _, value := range invalid {
		if IsSnakeIdentifier(value) {
			t.Fatalf("expected invalid snake identifier: %q", value)
		}
	}
}
