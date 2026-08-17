package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Selections struct {
	Theme       string `json:"theme"`
	Border      string `json:"border"`
	Density     string `json:"density"`
	Autoscroll  string `json:"autoscroll"`
	ScrollSpeed string `json:"scroll_speed"`
	Diagrams    string `json:"diagrams"`
}

func Defaults() Selections {
	return Selections{
		Theme:       "amber",
		Border:      "ascii",
		Density:     "normal",
		Autoscroll:  "off",
		ScrollSpeed: "medium",
		Diagrams:    "off",
	}
}

func Dir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "geetard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "geetard")
}

func Load(dir string) (Selections, string, error) {
	def := Defaults()
	b, err := os.ReadFile(filepath.Join(dir, "selections.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return def, "", nil
		}
		return def, "", err
	}
	var s Selections
	if json.Unmarshal(b, &s) != nil {
		return def, "bad config, using defaults", nil
	}
	return applyDefaults(s), "", nil
}

func applyDefaults(s Selections) Selections {
	d := Defaults()
	if s.Theme == "" {
		s.Theme = d.Theme
	}
	if s.Border == "" {
		s.Border = d.Border
	}
	if s.Density == "" {
		s.Density = d.Density
	}
	if s.Autoscroll == "" {
		s.Autoscroll = d.Autoscroll
	}
	if s.ScrollSpeed == "" {
		s.ScrollSpeed = d.ScrollSpeed
	}
	if s.Diagrams == "" {
		s.Diagrams = d.Diagrams
	}
	return s
}

func Save(dir string, s Selections) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "selections.json"), append(b, '\n'), 0o644)
}

func IntervalMs(speed string) int {
	switch speed {
	case "slow":
		return 800
	case "fast":
		return 200
	default:
		return 400
	}
}

func DensityGap(density string) int {
	switch density {
	case "compact":
		return 0
	case "airy":
		return 2
	default:
		return 1
	}
}
