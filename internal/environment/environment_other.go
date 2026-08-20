//go:build !darwin

package environment

import (
	"fmt"
	"os"
)

func Check(string, *os.File) error { return fmt.Errorf("init supports macOS on Apple silicon only") }
