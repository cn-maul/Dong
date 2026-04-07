//go:build webui
// +build webui

package main

import (
	"embed"
	"io/fs"
)

//go:embed web/*
var embeddedWeb embed.FS

func webUIFS() (fs.FS, bool) {
	return embeddedWeb, true
}
