package web

import "embed"

// Assets contains the production Svelte app and tracker. The release build
// copies both outputs into this directory before compiling the Go binary.
//
//go:embed all:dist
var Assets embed.FS
