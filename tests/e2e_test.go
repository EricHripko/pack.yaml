package e2e

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnd2End(t *testing.T) {
	// Arrange
	files, err := ioutil.ReadDir(".")
	if err != nil {
		t.Fatal("cannot discover tests")
	}

	// Act
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			// Arrange
			var stderr bytes.Buffer
			var stdout bytes.Buffer
			cmd := exec.Command(
				"docker",
				"build",
				"-t",
				file.Name(),
				"-f", "pack.yaml",
				".",
			)
			cmd.Dir = filepath.Join(".", file.Name())
			cmd.Stderr = &stderr
			cmd.Stdout = &stdout
			cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

			// Act
			err = cmd.Run()

			// Assert
			if err != nil {
				t.Log("stdout: ", stdout.String())
				t.Log("stderr: ", stderr.String())
				t.Log(err)
				t.Fail()
			}
		})
	}
}
