package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func build(t *testing.T, project string) {
	// Arrange
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd := exec.Command(
		"docker",
		"build",
		"-t",
		project,
		"-f", "pack.yaml",
		".",
	)
	cmd.Dir = filepath.Join(".", project)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

	// Act
	err := cmd.Run()

	// Assert
	if err != nil {
		t.Log("stdout: ", stdout.String())
		t.Log("stderr: ", stderr.String())
		t.Log(err)
		t.Fail()
	}
}

func TestGoCli(t *testing.T) {
	// Arrange
	project := "go-cli-template-master"
	build(t, project)

	var stdout bytes.Buffer
	cmd := exec.Command("docker", "run", "--rm", project)
	cmd.Stdout = &stdout

	// Act
	err := cmd.Run()

	// Assert
	require.Nil(t, err)
	require.Contains(t, stdout.String(), "cli-template")
}
