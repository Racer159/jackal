// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"slices"

	"github.com/defenseunicorns/zarf/src/pkg/utils/exec"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/stretchr/testify/require"
)

// ZarfE2ETest Struct holding common fields most of the tests will utilize.
type ZarfE2ETest struct {
	ZarfBinPath       string
	Arch              string
	ApplianceMode     bool
	ApplianceModeKeep bool
	RunClusterTests   bool
}

var logRegex = regexp.MustCompile(`Saving log file to (?P<logFile>.*?\.log)`)

// GetCLIName looks at the OS and CPU architecture to determine which Zarf binary needs to be run.
func GetCLIName() string {
	var binaryName string
	if runtime.GOOS == "linux" {
		binaryName = "zarf"
	} else if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			binaryName = "zarf-mac-apple"
		} else {
			binaryName = "zarf-mac-intel"
		}
	} else if runtime.GOOS == "windows" {
		if runtime.GOARCH == "amd64" {
			binaryName = "zarf.exe"
		}
	}
	return binaryName
}

// SetupWithCluster performs actions for each test that requires a K8s cluster.
func (e2e *ZarfE2ETest) SetupWithCluster(t *testing.T) {
	if !e2e.RunClusterTests {
		t.Skip("")
	}
	_ = exec.CmdWithPrint("sh", "-c", fmt.Sprintf("%s tools kubectl describe nodes | grep -A 99 Non-terminated", e2e.ZarfBinPath))
}

// Zarf executes a Zarf command.
func (e2e *ZarfE2ETest) Zarf(args ...string) (string, string, error) {
	if !slices.Contains(args, "--tmpdir") && !slices.Contains(args, "tools") {
		tmpdir, err := os.MkdirTemp("", "zarf-")
		if err != nil {
			return "", "", err
		}
		defer os.RemoveAll(tmpdir)
		args = append(args, "--tmpdir", tmpdir)
	}
	return exec.CmdWithContext(context.TODO(), exec.PrintCfg(), e2e.ZarfBinPath, args...)
}

// Kubectl executes `zarf tools kubectl ...`
func (e2e *ZarfE2ETest) Kubectl(args ...string) (string, string, error) {
	tk := []string{"tools", "kubectl"}
	args = append(tk, args...)
	return e2e.Zarf(args...)
}

// CleanFiles removes files and directories that have been created during the test.
func (e2e *ZarfE2ETest) CleanFiles(files ...string) {
	for _, file := range files {
		_ = os.RemoveAll(file)
	}
}

// GetMismatchedArch determines what architecture our tests are running on,
// and returns the opposite architecture.
func (e2e *ZarfE2ETest) GetMismatchedArch() string {
	switch e2e.Arch {
	case "arm64":
		return "amd64"
	default:
		return "arm64"
	}
}

// GetLogFileContents gets the log file contents from a given run's std error.
func (e2e *ZarfE2ETest) GetLogFileContents(t *testing.T, stdErr string) string {
	get, err := helpers.MatchRegex(logRegex, stdErr)
	require.NoError(t, err)
	logFile := get("logFile")
	logContents, err := os.ReadFile(logFile)
	require.NoError(t, err)
	return string(logContents)
}

// SetupDockerRegistry uses the host machine's docker daemon to spin up a local registry for testing purposes.
func (e2e *ZarfE2ETest) SetupDockerRegistry(t *testing.T, port int) {
	// spin up a local registry
	registryImage := "registry:2.8.3"
	err := exec.CmdWithPrint("docker", "run", "-d", "--restart=always", "-p", fmt.Sprintf("%d:5000", port), "--name", fmt.Sprintf("registry-%d", port), registryImage)
	require.NoError(t, err)
}

// TeardownRegistry removes the local registry.
func (e2e *ZarfE2ETest) TeardownRegistry(t *testing.T, port int) {
	// remove the local registry
	err := exec.CmdWithPrint("docker", "rm", "-f", fmt.Sprintf("registry-%d", port))
	require.NoError(t, err)
}

// GetZarfVersion returns the current build/zarf version
func (e2e *ZarfE2ETest) GetZarfVersion(t *testing.T) string {
	// Get the version of the CLI
	stdOut, stdErr, err := e2e.Zarf("version")
	require.NoError(t, err, stdOut, stdErr)
	return strings.Trim(stdOut, "\n")
}

// StripMessageFormatting strips any ANSI color codes and extra spaces from a given string
func (e2e *ZarfE2ETest) StripMessageFormatting(input string) string {
	// Regex to strip any color codes from the output - https://regex101.com/r/YFyIwC/2
	ansiRegex := regexp.MustCompile(`\x1b\[(.*?)m`)
	unAnsiInput := ansiRegex.ReplaceAllString(input, "")
	// Regex to strip any more than two spaces or newline - https://regex101.com/r/wqQmys/1
	multiSpaceRegex := regexp.MustCompile(`\s{2,}|\n`)
	return multiSpaceRegex.ReplaceAllString(unAnsiInput, " ")
}

// NormalizeYAMLFilenames normalizes YAML filenames / paths across Operating Systems (i.e Windows vs Linux)
func (e2e *ZarfE2ETest) NormalizeYAMLFilenames(input string) string {
	if runtime.GOOS != "windows" {
		return input
	}

	// Match YAML lines that have files in them https://regex101.com/r/C78kRD/1
	fileMatcher := regexp.MustCompile(`^(?P<start>.* )(?P<file>[^:\n]+\/.*)$`)
	scanner := bufio.NewScanner(strings.NewReader(input))

	output := ""
	for scanner.Scan() {
		line := scanner.Text()
		get, err := helpers.MatchRegex(fileMatcher, line)
		if err != nil {
			output += line + "\n"
			continue
		}
		output += fmt.Sprintf("%s\"%s\"\n", get("start"), strings.ReplaceAll(get("file"), "/", "\\\\"))
	}

	return output
}
