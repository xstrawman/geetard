package tui

import "geetard/internal/client"

func linePlainWidth(line client.DisplayLine) int {
	n := 0
	for _, p := range line.Parts {
		n += len([]rune(p.Text))
	}
	return n
}

func maxLineWidth(lines []client.DisplayLine) int {
	max := 0
	for _, ln := range lines {
		if w := linePlainWidth(ln); w > max {
			max = w
		}
	}
	return max
}

// sheetColumns is 1 for wide ASCII tabs, otherwise 1–3 newspaper columns
// from available width. Chord sheets (short lines) get more columns.
func sheetColumns(availWidth int, lines []client.DisplayLine) int {
	if availWidth < 72 {
		return 1
	}
	max := maxLineWidth(lines)
	if max == 0 {
		max = 40
	}
	if max >= 58 || max > availWidth/2 {
		return 1
	}
	colW := max + 2
	if colW < 32 {
		colW = 32
	}
	if colW > 56 {
		colW = 56
	}
	n := availWidth / (colW + 2)
	if n < 1 {
		n = 1
	}
	if n > 3 {
		n = 3
	}
	return n
}
