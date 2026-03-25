//go:build windows

// updater-helper is a tiny Windows-only binary.
// It is spawned by the updater before the main process exits.
// Args: <old-path> <new-path>
//   old-path — current executable (to be replaced)
//   new-path — downloaded .new file (replacement)
//
// Sequence:
//  1. Sleep 1.5s so the parent process can exit and release the file lock.
//  2. MoveFileEx(new-path → old-path, MOVEFILE_REPLACE_EXISTING).
//  3. exec.Command(old-path).Start() — re-launch with original args.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

var (
	modkernel32     = syscall.NewLazyDLL("kernel32.dll")
	procMoveFileExW = modkernel32.NewProc("MoveFileExW")
)

const moveFileReplaceExisting = 0x1

func moveFileEx(from, to string) error {
	lpFrom, err := syscall.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	lpTo, err := syscall.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	r1, _, e := procMoveFileExW.Call(
		uintptr(unsafe.Pointer(lpFrom)),
		uintptr(unsafe.Pointer(lpTo)),
		moveFileReplaceExisting,
	)
	if r1 == 0 {
		return e
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: updater-helper <old-exe> <new-exe>")
		os.Exit(1)
	}
	oldExe := os.Args[1]
	newExe := os.Args[2]

	// Wait for parent to exit and release the file lock.
	time.Sleep(1500 * time.Millisecond)

	if err := moveFileEx(newExe, oldExe); err != nil {
		fmt.Fprintf(os.Stderr, "updater-helper: MoveFileEx failed: %v\n", err)
		os.Exit(1)
	}

	// Re-launch the updated binary.
	cmd := exec.Command(oldExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "updater-helper: failed to start updated binary: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
