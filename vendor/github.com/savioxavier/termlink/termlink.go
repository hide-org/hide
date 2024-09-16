// Package termlink implements a set of functions to create customizable, clickable hyperlinks in the terminal.
package termlink

import (
	"fmt"
	"os"
	"strings"
)

// EnvironmentVariables represent the set of standalone environment variables
// ie, those which do not require any special handling or version checking
var EnvironmentVariables = []string{
	"DOMTERM",
	"WT_SESSION",
	"KONSOLE_VERSION",
}

// Version struct represents a semver version (usually, with some exceptions)
// with major, minor, and patch segments
type Version struct {
	major int
	minor int
	patch int
}

// parseVersion takes a string "version" number and returns
// a Version struct with the major, minor, and patch
// segments parsed from the string.
// If a version number is not provided
func parseVersion(version string) Version {
	var major, minor, patch int
	fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	return Version{
		major: major,
		minor: minor,
		patch: patch,
	}
}

// hasEnvironmentVariables returns true if the environment variable "name"
// is present in the environment, false otherwise
func hasEnv(name string) bool {
	_, envExists := os.LookupEnv(name)

	return envExists
}

// checkAllEnvs returns true if any of the environment variables in the "vars"
// string slice are actually present in the environment, false otherwise
func checkAllEnvs(vars []string) bool {
	for _, v := range vars {
		if hasEnv(v) {
			return true
		}
	}

	return false
}

// getEnv returns the value of the environment variable, if it exists
func getEnv(name string) string {
	envValue, _ := os.LookupEnv(name)

	return envValue
}

// matchesEnv returns true if the environment variable "name" matches any
// of the given values in the "values" string slice, false otherwise
func matchesEnv(name string, values []string) bool {
	if hasEnv(name) {
		for _, value := range values {
			if getEnv(name) == value {
				return true
			}
		}
	}
	return false
}

func supportsHyperlinks() bool {
	// Allow hyperlinks to be forced, independent of any environment variables
	// Instead of checking whether it is equal to anything other than "0",
	// a set of allowed values are provided, as something like
	// FORCE_HYPERLINK="do-not-enable-it" wouldn't make sense if it returned true
	if matchesEnv("FORCE_HYPERLINK", []string{"1", "true", "always", "enabled"}) {
		return true
	}

	// VTE-based terminals (Gnome Terminal, Guake, ROXTerm, etc)
	// VTE_VERSION is rendered as four-digit version string
	// eg: 0.52.2 => 5202
	// parseVersion will parse it with a standalone major segment
	// with minor and patch segments set to 0
	// 0.50.0 (parsed as 5000) was supposed to support hyperlinks, but throws a segfault
	// so we check if the "major" version is greater than 5000 (5000 exclusive)
	if hasEnv("VTE_VERSION") {
		v := parseVersion(getEnv("VTE_VERSION"))
		return v.major > 5000
	}

	// Terminals which have a TERM_PROGRAM variable set
	// This is the most versatile environment variable as it also provides another
	// variable called TERM_PROGRAM_VERSION, which helps us to determine
	// the exact version of the program, and allow for stricter variable checks
	if hasEnv("TERM_PROGRAM") {
		v := parseVersion(getEnv("TERM_PROGRAM_VERSION"))

		switch term := getEnv("TERM_PROGRAM"); term {
		case "iTerm.app":
			if v.major == 3 {
				return v.minor >= 1
			}
			return v.major > 3
		case "WezTerm":
			// Even though WezTerm's version is something like 20200620-160318-e00b076c
			// parseVersion will still parse it with a standalone major segment (ie: 20200620)
			// with minor and patch segments set to 0
			return v.major >= 20200620
		case "vscode":
			return v.major > 1 || (v.major == 1 && v.minor >= 72)

			// Hyper Terminal used to be included in this list, and it even supports hyperlinks
			// but the hyperlinks are pseudo-hyperlinks and are actually not clickable
		}
	}

	// Terminals which have a TERM variable set
	if matchesEnv("TERM", []string{"xterm-kitty", "alacritty", "alacritty-direct"}) {
		return true
	}

	// Terminals which have a COLORTERM variable set
	if matchesEnv("COLORTERM", []string{"xfce4-terminal"}) {
		return true
	}

	// Terminals in JetBrains IDEs
	if matchesEnv("TERMINAL_EMULATOR", []string{"JetBrains-JediTerm"}) {
		return true
	}

	// Match standalone environment variables
	// ie, those which do not require any special handling
	// or version checking
	if checkAllEnvs(EnvironmentVariables) {
		return true
	}

	return false
}

var colorsList = map[string]int{
	"reset":     0,
	"bold":      1,
	"dim":       2,
	"italic":    3,
	"underline": 4,
	"blink":     5,
	"black":     30,
	"red":       31,
	"green":     32,
	"yellow":    33,
	"blue":      34,
	"magenta":   35,
	"cyan":      36,
	"white":     37,
	"bgBlack":   40,
	"bgRed":     41,
	"bgGreen":   42,
	"bgYellow":  43,
	"bgBlue":    44,
	"bgMagenta": 45,
	"bgCyan":    46,
	"bgWhite":   47,
}

// isInList returns true if a given value is present in a string slice, false otherwise
func isInList(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

var colors []string

// addColor adds a color to be later used while parsing the color string
func addColor(value string) []string {
	colors = append(colors, fmt.Sprint(colorsList[value]))

	return colors
}

// clearSlice clears a string slice by re-slicing it to have a length of 0
func clearSlice(slice []string) []string {
	slice = slice[:0]
	return slice
}

// isValidColor checks if a given color string is a valid key in a map of predefined colors
func isValidColor(color string) bool {
	// Create a slice with keys of the colorsList map
	keys := make([]string, len(colorsList))

	i := 0
	for k := range colorsList {
		keys[i] = k
		i++
	}

	// Check if the color is in the keys slice
	return isInList(keys, color)
}

// parseColor yields a set of ANSI color codes, based on the "color" string.
// The color codes are parsed from the colorsList map
func parseColor(color string) string {
	// Clear the colors slice first
	colors = clearSlice(colors)

	// If nothing is provided, return empty string
	if color == "" {
		return ""
	}

	colorNames := strings.Split(color, " ")

	for _, c := range colorNames {
		// If the color doesn't exist, skip and go to the next word
		if !isValidColor(c) {
			continue
		}

		// Add the color, if present in colorsList
		addColor(c)

	}

	return "\u001b[" + strings.Join(colors, ";") + "m"
}

// Link returns a clickable link, which can be used in the terminal
//
// The function takes two required parameters: text and url
// and one optional parameter: shouldForce
//
// The text parameter is the text to be displayed.
// The url parameter is the URL to be opened when the link is clicked.
// The shouldForce is an optional parameter indicates whether to force the non-hyperlink supported behavior (i.e., text (url))
//
// The function returns the clickable link.
func Link(text string, url string, shouldForce ...bool) string {

	// Default shouldForce to false
	shouldForceDefault := false

	if len(shouldForce) > 0 {
		// If a value for shouldForce is provided, set it to that instead
		// Since shouldForce is a slice, we only consider its first element
		shouldForceDefault = shouldForce[0]
	}

	if shouldForceDefault {
		return text + " (" + url + ")" + "\u001b[0m"
	} else {
		if supportsHyperlinks() {
			return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07" + "\u001b[0m"
		}
		return text + " (" + url + ")" + "\u001b[0m"
	}
}

// ColorLink returns a colored clickable link, which can be used in the terminal
//
// The function takes three required parameters: text, url and color
// and one optional parameter: shouldForce
//
// The text parameter is the text to be displayed.
// The url parameter is the URL to be opened when the link is clicked.
// The color parameter is the color of the link.
// The shouldForce is an optional parameter indicates whether to force the non-hyperlink supported behavior (i.e., text (url))
//
// The function returns the clickable link.
func ColorLink(text string, url string, color string, shouldForce ...bool) string {
	textColor := parseColor(color)

	// Default shouldForce to false
	shouldForceDefault := false

	if len(shouldForce) > 0 {
		// If a value for shouldForce is provided, set it to that instead
		// Since shouldForce is a slice, we only consider its first element
		shouldForceDefault = shouldForce[0]
	}

	if shouldForceDefault {
		return textColor + text + " (" + url + ")" + "\u001b[0m"
	} else {
		if supportsHyperlinks() {
			return "\x1b]8;;" + url + "\x07" + textColor + text + "\x1b]8;;\x07" + "\u001b[0m"
		}
		return textColor + text + " (" + url + ")" + "\u001b[0m"
	}

}

// SupportsHyperlinks returns true if the terminal supports hyperlinks.
//
// The function returns true if the terminal supports hyperlinks, false otherwise.
func SupportsHyperlinks() bool {
	return supportsHyperlinks()
}
