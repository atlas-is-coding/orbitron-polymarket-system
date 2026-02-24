//go:build tools

package main

import (
	_ "github.com/fsnotify/fsnotify"
	_ "github.com/google/uuid"
	_ "modernc.org/sqlite"
)
