package portable

import (
	_ "embed"
	"errors"
	"strings"
)

// CapabilityGuide ships with the binary, independent of source checkout/cwd.
//
//go:embed guides/capability-authoring.md
var CapabilityGuide string

func CapabilityTemplate(id, description string) (Capability, error) {
	if !identifier.MatchString(id) {
		return Capability{}, errors.New("--capability must be 3-128 lowercase letters, digits, or hyphens, starting with a letter; see agent capability-guide")
	}
	if strings.TrimSpace(description) == "" {
		description = "Describe when to use this capability and the outcome it provides."
	}
	return Capability{Version: 1, ID: id, Description: description, Subscriptions: []Subscription{}, Checks: [][]string{{"sh", "checks/check.sh"}}}, nil
}
