package client

import "fmt"

var stringNames = []string{"e", "B", "G", "D", "A", "E"}

func Shapes(tabView map[string]any) []ChordShape {
	app, _ := tabView["applicature"].(map[string]any)
	if app == nil {
		return nil
	}
	var out []ChordShape
	for name, raw := range app {
		variants, _ := raw.([]any)
		if len(variants) == 0 {
			continue
		}
		first, _ := variants[0].(map[string]any)
		if first == nil {
			continue
		}
		fretsAny, _ := first["frets"].([]any)
		if len(fretsAny) < 6 {
			continue
		}
		lines := make([]string, 6)
		for i := 0; i < 6; i++ {
			fret := AsInt(fretsAny[5-i], 0)
			cell := "0"
			if fret < 0 {
				cell = "x"
			} else {
				cell = fmt.Sprintf("%d", fret)
			}
			lines[i] = fmt.Sprintf("  %s |-%s-", stringNames[i], cell)
		}
		out = append(out, ChordShape{Name: name, Lines: lines})
	}
	return out
}
