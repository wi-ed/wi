// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package colors declare constants and functions to simplify color management.
package colors

// Known colors.
var (
	Black         = RGB{0, 0, 0}
	Blue          = RGB{0, 0, 170}
	Green         = RGB{0, 170, 0}
	Cyan          = RGB{0, 170, 170}
	Red           = RGB{170, 0, 0}
	Magenta       = RGB{170, 0, 170}
	Brown         = RGB{170, 85, 0}
	LightGray     = RGB{170, 170, 170}
	DarkGray      = RGB{85, 85, 85}
	BrightBlue    = RGB{85, 85, 255}
	BrightGreen   = RGB{85, 255, 85}
	BrightCyan    = RGB{85, 255, 255}
	BrightRed     = RGB{255, 85, 85}
	BrightMagenta = RGB{255, 85, 255}
	BrightYellow  = RGB{255, 255, 85}
	White         = RGB{255, 255, 255}
)

// EGA lists the colors to use for maximum compatibility with terminals that
// do not support terminal-256.
//
// EGA colors <3 forever.
// Random intertubes websites says things like:
//   sudo apt-get install ncurses-term
//   export TERM=xterm-256color
//
// "tput colors" can be used to determine the number of colors supported by
// the terminal.
var EGA = []RGB{
	Black,
	Blue,
	Green,
	Cyan,
	Red,
	Magenta,
	Brown,
	LightGray,
	DarkGray,
	BrightBlue,
	BrightGreen,
	BrightCyan,
	BrightRed,
	BrightMagenta,
	BrightYellow,
	White,
}

// TODO(maruel): Add all colors for VGA;
// gnome-256color/putty-256color/rxvt-256color/xterm-256color.
// Debian package ncurses-term gets the DB for colors.

// RGB represents the color of a single character on screen.
//
// Transparency per character is not supported, this is text mode after all.
type RGB struct {
	R, G, B uint8
}

// NearestEGA returns the nearest colors for a 16 colors terminal.
func NearestEGA(c RGB) RGB {
	minDistance := 255 * 255 * 3
	out := Black
	for _, ega := range EGA {
		r := int(ega.R) - int(c.R)
		g := int(ega.G) - int(c.G)
		b := int(ega.B) - int(c.B)
		distance := r*r + b*b + g*g
		if distance < minDistance {
			minDistance = distance
			out = ega
		}
	}
	return out
}
