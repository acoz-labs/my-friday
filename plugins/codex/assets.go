// Package codexplugin carries the native plugin in the executable so installation
// does not depend on a source checkout or a separately downloaded moving branch.
package codexplugin

import "embed"

// Files contains only the public plugin package and marketplace, never a bank
// or native authentication. all: includes the dot-prefixed native manifests.
//
//go:embed all:plugins all:.agents
var Files embed.FS
