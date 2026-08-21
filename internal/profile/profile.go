package profile

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
	"golang.org/x/text/unicode/norm"
)

type Profile struct {
	ContractVersion int           `json:"contract_version"`
	AssistantID     string        `json:"assistant_id"`
	Identity        Identity      `json:"identity"`
	Communication   Communication `json:"communication"`
}
type Identity struct {
	DisplayName   string  `json:"display_name"`
	AddressUserAs *string `json:"address_user_as"`
	Purpose       string  `json:"purpose"`
}
type Communication struct {
	Preset         string  `json:"preset"`
	CustomGuidance *string `json:"custom_guidance"`
}

func Normalize(value string, limit int, required bool) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("text must be valid UTF-8")
	}
	value = norm.NFC.String(strings.TrimFunc(value, unicode.IsSpace))
	if required && value == "" {
		return "", fmt.Errorf("text is required")
	}
	for _, r := range value {
		if r == 0 || unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp) {
			return "", fmt.Errorf("text contains a prohibited control or format character")
		}
	}
	if uniseg.GraphemeClusterCount(value) > limit {
		return "", fmt.Errorf("text exceeds %d visible characters", limit)
	}
	return value, nil
}

func New(name, address, purpose, preset, guidance string) (Profile, error) {
	var p Profile
	var err error
	p.ContractVersion = 1
	if p.Identity.DisplayName, err = Normalize(name, 60, true); err != nil {
		return p, fmt.Errorf("display name: %w", err)
	}
	address, err = Normalize(address, 60, false)
	if err != nil {
		return p, fmt.Errorf("form of address: %w", err)
	}
	if address != "" {
		p.Identity.AddressUserAs = &address
	}
	if p.Identity.Purpose, err = Normalize(purpose, 240, true); err != nil {
		return p, fmt.Errorf("purpose: %w", err)
	}
	preset = strings.ToLower(strings.TrimSpace(preset))
	switch preset {
	case "balanced", "concise", "conversational", "custom":
	default:
		return p, fmt.Errorf("communication preset must be balanced, concise, conversational, or custom")
	}
	p.Communication.Preset = preset
	guidance, err = Normalize(guidance, 240, preset == "custom")
	if err != nil {
		return p, fmt.Errorf("custom guidance: %w", err)
	}
	if preset != "custom" && guidance != "" {
		return p, fmt.Errorf("custom guidance is only valid with the custom preset")
	}
	if guidance != "" {
		p.Communication.CustomGuidance = &guidance
	}
	return p, nil
}

// Validate verifies a decoded profile without collapsing null and empty.
func Validate(p Profile) error {
	if p.ContractVersion != 1 {
		return fmt.Errorf("unsupported profile contract version")
	}
	name, err := Normalize(p.Identity.DisplayName, 60, true)
	if err != nil || name != p.Identity.DisplayName {
		return fmt.Errorf("display name is not canonically normalized")
	}
	purpose, err := Normalize(p.Identity.Purpose, 240, true)
	if err != nil || purpose != p.Identity.Purpose {
		return fmt.Errorf("purpose is not canonically normalized")
	}
	if p.Identity.AddressUserAs != nil {
		if *p.Identity.AddressUserAs == "" {
			return fmt.Errorf("form of address must be null or non-empty")
		}
		address, e := Normalize(*p.Identity.AddressUserAs, 60, false)
		if e != nil || address != *p.Identity.AddressUserAs {
			return fmt.Errorf("form of address is not canonically normalized")
		}
	}
	switch p.Communication.Preset {
	case "balanced", "concise", "conversational":
		if p.Communication.CustomGuidance != nil {
			return fmt.Errorf("custom guidance must be null unless preset is custom")
		}
	case "custom":
		if p.Communication.CustomGuidance == nil {
			return fmt.Errorf("custom guidance is required for custom preset")
		}
		guidance, e := Normalize(*p.Communication.CustomGuidance, 240, true)
		if e != nil || guidance != *p.Communication.CustomGuidance {
			return fmt.Errorf("custom guidance is not canonically normalized")
		}
	default:
		return fmt.Errorf("unknown communication preset")
	}
	return nil
}
