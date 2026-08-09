package main

import (
	"flag"
	"testing"
)

func TestParseFlags_flagsAfterPath(t *testing.T) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	lang := fs.String("lang", "", "lang")
	format := fs.String("format", "text", "format")

	rest, err := parseFlags(fs, []string{"./mod", "--lang", "ru", "--format", "json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 1 || rest[0] != "./mod" {
		t.Fatalf("rest = %#v", rest)
	}
	if *lang != "ru" || *format != "json" {
		t.Fatalf("lang=%q format=%q", *lang, *format)
	}
}

func TestParseFlags_equalsForm(t *testing.T) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	lang := fs.String("lang", "", "lang")

	rest, err := parseFlags(fs, []string{"--lang=ru", "path"})
	if err != nil {
		t.Fatal(err)
	}
	if rest[0] != "path" || *lang != "ru" {
		t.Fatalf("rest=%#v lang=%q", rest, *lang)
	}
}
