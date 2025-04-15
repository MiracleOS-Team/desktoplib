package foreignToplevel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	"github.com/MiracleOS-Team/libxdg-go/icons"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xwindow"
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

func listToplevelsXorg() ([]Toplevel, error) {
	var tpl []Toplevel

	X, err := xgbutil.NewConn()
	if err != nil {
		return tpl, err
	}

	// List top-level windows
	clientList, err := ewmh.ClientListGet(X)
	if err != nil {
		return tpl, err
	}

	for _, win := range clientList {
		name, _ := ewmh.WmNameGet(X, win)
		class, _ := icccm.WmClassGet(X, win)

		ntpl := Toplevel{
			Title: name,
			AppID: class.Class,
		}

		tpl = append(tpl, ntpl)
	}

	return tpl, nil
}

// ListToplevels lists all toplevel windows and parses the output into Toplevel structs
func ListToplevels() ([]Toplevel, error) {
	output, err := runWlrctlCommand([]string{"toplevel", "list"})
	if err != nil {
		// Fallback to Xorg support
		return listToplevelsXorg()
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

func selectToplevelXorg(toplevel Toplevel) error {
	X, err := xgbutil.NewConn()
	if err != nil {
		return err
	}

	clientList, err := ewmh.ClientListGet(X)
	if err != nil {
		return err
	}

	var matchedWin xproto.Window
	for _, win := range clientList {
		title, _ := ewmh.WmNameGet(X, win)
		class, _ := icccm.WmClassGet(X, win)

		if strings.Contains(strings.ToLower(title), strings.ToLower(toplevel.Title)) &&
			strings.EqualFold(class.Class, toplevel.AppID) {

			matchedWin = win
			break
		}
	}

	if matchedWin == 0 {
		return fmt.Errorf("no matching window")
	}

	xw := xwindow.New(X, matchedWin)
	xw.Focus()

	err = xproto.ConfigureWindowChecked(X.Conn(), matchedWin, xproto.ConfigWindowStackMode, []uint32{xproto.StackModeAbove}).Check()
	if err != nil {
		return err
	}

	return nil
}

// SelectToplevel selects a toplevel window based on a match specifier
func SelectToplevel(toplevel Toplevel) error {
	var matchSpecs []string = generateMatchSpecifiers(toplevel)
	// Focus the toplevel matching the specifier
	_, err := runWlrctlCommand(append([]string{"toplevel", "focus"}, matchSpecs...))
	if err != nil {
		return selectToplevelXorg(toplevel)
	}

	return nil
}

// loadRules reads the JSON file and returns a map of regex patterns to replacements.
func loadAliases(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var rules map[string]string
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func GetIconFromToplevel(toplevel Toplevel, size int, scale int) (string, error) {
	icon_name := toplevel.AppID

	rules, err := loadAliases("/opt/miracleos-software/desk-data/app-icons-alias.json")
	if err == nil {
		for pattern, replacement := range rules {
			// Remove the `/g` flag if present in the JSON file.
			cleanPattern := pattern
			if pattern[0] == '/' && pattern[len(pattern)-2:] == "/g" {
				cleanPattern = pattern[1 : len(pattern)-2]
			}

			re := regexp.MustCompile(cleanPattern)
			if re.MatchString(icon_name) {
				icon_name = replacement
			}
		}
	} else {
		return "", err
	}

	ic, err := icons.FindIconDefaults(icon_name, size, scale, "application-x-executable")
	if strings.Contains(ic, "application-x-executable") {
		ic, err = icons.FindIconDefaults(strings.ToLower(icon_name), size, scale, "application-x-executable")

	}

	return ic, err
}
