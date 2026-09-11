package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	s := Defaults()
	if s.Theme != "amber" || s.Border != "ascii" || s.Density != "normal" ||
		s.Autoscroll != "off" || s.ScrollSpeed != "medium" || s.Diagrams != "off" {
		t.Fatalf("%+v", s)
	}
}

func TestLoad_ok(t *testing.T) {
	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "selections_ok.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "selections.json"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	s, warn, err := Load(dir)
	if err != nil || warn != "" {
		t.Fatal(err, warn)
	}
	if s.Theme != "nord" || s.Diagrams != "sidebar" || s.Autoscroll != "on" {
		t.Fatalf("%+v", s)
	}
}

func TestLoad_missingUsesDefaults(t *testing.T) {
	s, warn, err := Load(t.TempDir())
	if err != nil || warn != "" || s.Theme != "amber" {
		t.Fatal(s, warn, err)
	}
}

func TestLoad_corrupt(t *testing.T) {
	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "selections_bad.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "selections.json"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	s, warn, err := Load(dir)
	if err != nil || warn != "bad config, using defaults" || s.Theme != "amber" {
		t.Fatal(s, warn, err)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := Defaults()
	in.Theme = "green"
	if err := Save(dir, in); err != nil {
		t.Fatal(err)
	}
	out, warn, err := Load(dir)
	if err != nil || warn != "" || out.Theme != "green" {
		t.Fatal(out, warn, err)
	}
}

func TestIntervalMs(t *testing.T) {
	if IntervalMs("slow") != 800 || IntervalMs("medium") != 400 || IntervalMs("fast") != 200 {
		t.Fatal(IntervalMs("slow"), IntervalMs("medium"), IntervalMs("fast"))
	}
}

func TestLoad_unknownValuesFallBack(t *testing.T) {
	dir := t.TempDir()
	src := []byte(`{"theme":"hotpink","border":"neon","density":"huge","autoscroll":"maybe","scroll_speed":"ludicrous","diagrams":"3d"}`)
	if err := os.WriteFile(filepath.Join(dir, "selections.json"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	s, _, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	d := Defaults()
	if s != d {
		t.Fatalf("got %+v want %+v", s, d)
	}
}
