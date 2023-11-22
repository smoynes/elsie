// Package tool defines very naive scripts for development tasks. These are not intended to be
// portable and simply replace rote commands with tasks. Think of them as executable screenplays.
// Just like shell, it is a miracle when these scripts work at all.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	path "path/filepath"
	"runtime"
	"strings"
)

var usage = `go run internal/tool <COMMAND>

Commands:

- deps       checks build dependencies: (stringer, docker, golangci-lint)
- container  builds docker image: smoynes/elsie
- lint       check style with golangci-lint
`

func main() {
	args := os.Args

	if err := projectWorkingDirectory(); err != nil {
		log.Fatal(err)
	}

	switch {
	case len(args) == 2 && os.Args[1] == "deps":
		if err := deps(); err != nil {
			log.Fatal(err)
		}
	case len(args) == 2 && os.Args[1] == "container":
		if err := dockerBuild(); err != nil {
			log.Fatal(err)
		}
	case len(args) == 2 && os.Args[1] == "lint":
		if err := golangciLint(); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Usage: %s\n", usage)
	}
}

// projectWorkingDirectory finds the project directory and changes the working directory to it. The
// project directory is the working directory or its ancestor with a go.mod file. If a project
// directory is not found or, to prevent inadvertent catastrophes, it is found to be a root
// directory, an error is returned.
func projectWorkingDirectory() error {
	dir, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	for {
		file := path.Join(dir, "go.mod")

		if _, err := os.Stat(file); err == nil {
			break
		} else if os.IsNotExist(err) {
			dir = path.Dir(dir)
		} else {
			return err
		}
	}

	if dir == path.Dir(dir) {
		return errors.New("project directory is root directory")
	}

	if err := os.Chdir(dir); err != nil {
		return err
	}

	return nil
}

func deps() error {
	if stringer, err := exec.LookPath("stringer"); err != nil {
		return fmt.Errorf("stringer (required): %w", err)
	} else {
		fmt.Println("stringer (required):", stringer)
	}

	if linter, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint (optional): %w", err)
	} else {
		fmt.Println("golangci-lint (optional):", linter)
	}

	if docker, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker (optional): version: %w", err)
	} else {
		fmt.Println("docker (optional):", docker)
	}

	return nil
}

func dockerBuild() error {
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	//nolint:gosec
	docker := exec.Command("docker", "build",
		"-t", "smoynes/elsie",
		"--build-arg", "GOLANG_VERSION="+goVersion,
		".",
	)
	out, err := docker.StderrPipe()

	if err != nil {
		return fmt.Errorf("docker: pipe: %w", err)
	}

	if err = docker.Start(); err != nil {
		return fmt.Errorf("docker: build: %w", err)
	}

	fmt.Println("docker build:")

	for {
		copied, err := io.Copy(os.Stdout, out)
		if err != nil {
			return fmt.Errorf("docker: io: %w", err)
		}

		if copied == 0 {
			break
		}
	}

	if err = docker.Wait(); err != nil {
		return fmt.Errorf("docker: wait: %w", err)
	}

	println("\n\nBuilt container:")
	println("\tdocker run smoynes/elsie")

	return nil
}

func golangciLint() error {
	linter := exec.Command("golangci-lint", "run")
	out, err := linter.StdoutPipe()

	if err != nil {
		return fmt.Errorf("golangci-lint: pipe: %w", err)
	}

	if err = linter.Start(); err != nil {
		return fmt.Errorf("golangci-lint: run: %w", err)
	}

	fmt.Println("golangci-lint:")

	for {
		copied, err := io.Copy(os.Stdout, out)
		if err != nil {
			return fmt.Errorf("golangci-lint: io: %w", err)
		}

		if copied == 0 {
			break
		}
	}

	if err = linter.Wait(); err != nil {
		return fmt.Errorf("golangci-lint: wait: %w", err)
	}

	return nil
}
