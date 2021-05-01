package frontend

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/EricHripko/pack.yaml/pkg/packer2llb"

	"github.com/EricHripko/buildkit-fdk/pkg/cib"
	"github.com/moby/buildkit/frontend/gateway/client"
	fsutil "github.com/tonistiigi/fsutil/types"
)

// Returned when no command was found in the produced image.
var errNoCommand = errors.New("frontend: no command found")

// Looks in the known install directory and attempts to automatically detect
// the command for the image.
func findCommand(ctx context.Context, ref client.Reference) (command string, err error) {
	prefix := packer2llb.DirInstall[1:]
	err = cib.WalkRecursive(ctx, ref, func(file *fsutil.Stat) error {
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
			return errors.New("frontend: multiple commands found (" + command + ")")
		}
		command = "/" + file.Path
		return nil
	})
	if command == "" {
		err = errNoCommand
	}
	return
}
