// Package tool defines very naive scripts for development tasks. These are not intended to be
// portable and simply replace equivalent shell scripts. Just like shell, it is a miracle when these
// scripts work at all.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	path "path/filepath"
)

var usage = `go run internal/tool <COMMAND>

Commands:

- deps       checks build dependencies: (stringer, docker)
- container  builds docker image: smoynes/elsie
- lint       check style with golangci-lint
`

func main() {
	args := os.Args
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
			log.Fatal(err)
		}
	}

	if err := os.Chdir(dir); err != nil {
		log.Fatal(err)
	}

	switch {
	case len(args) == 2 && os.Args[1] == "deps":
		if err := installTools(); err != nil {
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

func installTools() error {
	if stringer, err := exec.LookPath("stringer"); err != nil {
		return fmt.Errorf("stringer: %w", err)
	} else {
		fmt.Println("stringer:", stringer)
	}

	docker := exec.Command("docker", "version")

	if err := docker.Run(); err != nil {
		return fmt.Errorf("docker: version: %w", err)
	} else {
		fmt.Println("docker:", docker)
	}

	return nil
}

func dockerBuild() error {
	docker := exec.Command("docker", "build", "-t", "smoynes/elsie", ".")
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
