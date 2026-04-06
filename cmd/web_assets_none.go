//go:build !webui
// +build !webui

package main

import "io/fs"

func webUIFS() (fs.FS, bool) {
	return nil, false
}
