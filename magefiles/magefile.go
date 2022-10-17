//go:build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/aserto-dev/mage-loot/common"
	"github.com/aserto-dev/mage-loot/deps"
)

const asertoCLI = "aserto"

// Lint runs linting for the entire project.
func Lint() error {
	return common.Lint()
}

// Test runs all tests and generates a code coverage report.
func Test() error {
	return common.Test()
}

func Deps() {
	deps.GetAllDeps()
}

// SetupExamples starts the aserto one-box with a sample policy used by middleware examples.
func SetupExamples() error {
	exPath, err := examplesPath()
	if err != nil {
		return errors.Wrap(err, "unable to locate examples directory")
	}

	installed, err := oneboxInstalled()
	if err != nil || !installed {
		if err := installOnebox(); err != nil {
			return errors.Wrap(err, "unable to install onebox")
		}
	}

	if err := startOnebox(filepath.Join(exPath, "policy")); err != nil {
		return err
	}

	if err := loadExampleUsers(filepath.Join(exPath, "users.json")); err != nil {
		fmt.Println("Failed to load sample users")
		return err
	}

	return nil
}

// TeardownExamples stops the aserto one-box environment started with SetupExamples.
func TeardownExamples() error {
	return runCLI("developer", "stop")
}

// ListExamples lists examples that can be run with the aserto one-box.
func ListExamples() error {
	exPath, err := examplesPath()
	if err != nil {
		return errors.Wrap(err, "unable to locate examples directory")
	}

	files, err := ioutil.ReadDir(filepath.Join(exPath, "http"))
	if err != nil {
		return errors.Wrapf(err, "unable to list examples under '%s'", exPath)
	}

	for _, file := range files {
		if file.IsDir() {
			fmt.Println(file.Name())
		}
	}

	return nil
}

// Example prints the command to start the specified example.
// You can run the example using $(mage example <name>).
// For example: $(mage example gin)
func Example(name string) error {
	exPath, err := examplesPath()
	if err != nil {
		return errors.Wrap(err, "unable to locate examples directory")
	}

	fmt.Println("go run", filepath.Join(exPath, "http", name))

	return nil
}

func examplesPath() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, "examples", "middleware"), nil
}

func oneboxInstalled() (bool, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return false, errors.Wrap(err, "unable to find home directory")
	}

	if _, err := os.Stat(filepath.Join(homedir, ".config", "aserto", "aserto-one", "certs")); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func installOnebox() error {
	return runCLI("developer", "install", "--trust-cert")
}

func startOnebox(policyPath string) error {
	fmt.Println("\nStarting aserto onebox with policy from", policyPath)
	return runCLI("developer", "start", "local", "--src-path", policyPath)
}

func loadExampleUsers(usersPath string) error {
	fmt.Println("\nLoading sample users...")
	return runCLI(
		"directory", "load-users",
		"--provider", "json", "--file", usersPath, "--incl-user-ext",
		"--authorizer", "localhost:8282",
	)
}

func runCLI(args ...string) error {
	err := runCommand(asertoCLI, args...)
	if err != nil && !errors.Is(err, &exec.ExitError{}) {
		// Command couldn't be started
		fmt.Printf("unable to start aserto onebox. make sure '%s' is installed\n", asertoCLI)
	}

	return err
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Command started and failed
			fmt.Println(string(exitErr.Stderr))
		}

		return err
	}

	return nil
}
