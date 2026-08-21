package environment

import (
	"fmt"
	"strconv"
	"strings"
)

func validateContract(architecture, macOSVersion, filesystem, gitVersion string, interactive bool) error {
	if architecture != "arm64" {
		return fmt.Errorf("unsupported architecture %s; Apple silicon is required", architecture)
	}
	major, _ := strconv.Atoi(strings.Split(macOSVersion, ".")[0])
	if major < 14 {
		return fmt.Errorf("macOS 14 or later is required; found %s", macOSVersion)
	}
	if !interactive {
		return fmt.Errorf("init requires an interactive terminal")
	}
	if filesystem != "apfs" {
		return fmt.Errorf("local APFS is required; found %s", filesystem)
	}
	var gitMajor, gitMinor int
	if _, err := fmt.Sscanf(strings.TrimSpace(gitVersion), "git version %d.%d", &gitMajor, &gitMinor); err != nil || gitMajor < 2 || gitMajor == 2 && gitMinor < 28 {
		return fmt.Errorf("Git 2.28 or later is required; found %s", strings.TrimSpace(gitVersion))
	}
	return nil
}
