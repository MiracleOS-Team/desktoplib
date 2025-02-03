package foreignToplevel

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Toplevel represents a toplevel window with relevant attributes
type Toplevel struct {
	AppID string
	Title string
	State string
}

// RunWlrctlCommand runs a wlrctl command and returns the output or an error
func runWlrctlCommand(args []string) (string, error) {
	cmd := exec.Command("wlrctl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing wlrctl: %v, stderr: %s, stdout: %s", err, stderr.String(), stdout.String())
	}

	return stdout.String(), nil
}

// ListToplevels lists all toplevel windows and parses the output into Toplevel structs
func ListToplevels() ([]Toplevel, error) {
	output, err := runWlrctlCommand([]string{"toplevel", "list"})
	if err != nil {
		return nil, err
	}

	// Parse the toplevel information into structs
	var toplevels []Toplevel
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split the line at the first colon to get AppID and Title
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			// If no colon, skip the line
			continue
		}

		appID := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])

		// For now, we assume state is "active" as a placeholder

		// Append the toplevel to the list
		toplevels = append(toplevels, Toplevel{
			AppID: appID,
			Title: title,
		})
	}

	return toplevels, nil
}

// GenerateMatchSpecifiers generates match specifiers for a given Toplevel
func generateMatchSpecifiers(toplevel Toplevel) []string {
	var matchSpecs []string

	// Generate match specifiers based on attributes
	if toplevel.AppID != "" {
		matchSpecs = append(matchSpecs, fmt.Sprintf("app_id:\"%s\"", toplevel.AppID))
	}
	if toplevel.Title != "" {
		matchSpecs = append(matchSpecs, fmt.Sprintf("title:\"%s\"", toplevel.Title))
	}
	if toplevel.State != "" {
		matchSpecs = append(matchSpecs, fmt.Sprintf("state:\"%s\"", toplevel.State))
	}

	// Join all match specifiers into a single string without spaces
	return matchSpecs
}

// SelectToplevel selects a toplevel window based on a match specifier
func SelectToplevel(toplevel Toplevel) error {
	var matchSpecs []string = generateMatchSpecifiers(toplevel)
	// Focus the toplevel matching the specifier
	_, err := runWlrctlCommand(append([]string{"toplevel", "focus"}, matchSpecs...))
	if err != nil {
		return fmt.Errorf("error selecting toplevel: %v", err)
	}

	return nil
}
