package cib

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/moby/buildkit/frontend/gateway/client"
	fsutil "github.com/tonistiigi/fsutil/types"
)

// ErrNoCommand is returned when no command was found in the produced image.
var ErrNoCommand = errors.New("cib: no command found")

// WalkFunc is the type of function called for each file or directory visited
// by WalkRecursive.
type WalkFunc func(file *fsutil.Stat) error

// WalkRecursive iterates all the files in the reference recursively.
func WalkRecursive(ctx context.Context, ref client.Reference, walkFn WalkFunc) error {
	return walkRecursive(ctx, ref, ".", walkFn)
}

func walkRecursive(ctx context.Context, ref client.Reference, root string, walkFn WalkFunc) error {
	files, err := ref.ReadDir(ctx, client.ReadDirRequest{Path: root})
	if err != nil {
		return err
	}

	for _, file := range files {
		// Make path absolute for easier integration
		file.Path = path.Join(root, file.Path)

		// Callback
		err = walkFn(file)
		if err != nil {
			return err
		}

		// Walk folders
		mode := os.FileMode(file.Mode)
		if mode.IsDir() {
			err = walkRecursive(ctx, ref, file.Path, walkFn)
			if err != nil {
				return err
			}
		}
	}
	return nil

}

// FindCommand looks in the known install directory and attempts to
// automatically detect the command for the image.
func FindCommand(ctx context.Context, ref client.Reference) (command string, err error) {
	prefix := DirInstall[1:]
	err = WalkRecursive(ctx, ref, func(file *fsutil.Stat) error {
		// Must be in install location
		if !strings.HasPrefix(file.Path, prefix) {
			return nil
		}
		// Must not be a directory
		if os.FileMode(file.Mode).IsDir() {
			return nil
		}
		// Must be executable
		if file.Mode&0100 == 0 {
			return nil
		}
		// Multiple commands found
		if command != "" {
			return errors.New("cib: multiple commands found (" + command + ")")
		}
		command = "/" + file.Path
		return nil
	})
	if command == "" {
		err = ErrNoCommand
	}
	return
}
