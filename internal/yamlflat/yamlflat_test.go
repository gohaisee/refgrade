package yamlflat

import "testing"

func TestParse_flat(t *testing.T) {
	t.Parallel()

	got, err := Parse([]byte(`# comment
lang: ru
profile: standard
key: "quoted"
`))
	if err != nil {
		t.Fatal(err)
	}
	if got["lang"] != "ru" || got["profile"] != "standard" || got["key"] != "quoted" {
		t.Fatalf("Parse = %#v", got)
	}
}

func TestParse_errors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
	}{
		{"no colon", "lang ru"},
		{"empty key", ": value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Parse([]byte(tc.in)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	t.Parallel()

	if unquote(`'ru'`) != "ru" {
		t.Fatal("single quotes")
	}
	if unquote("plain") != "plain" {
		t.Fatal("plain")
	}
}
