// Package tool defines very naive scripts for development tasks. These are not
// intended to be portable but instead simply replace rote commands with tasks.
// Think of them as executable screenplays. Just like shell, it is a miracle
// these scripts work at all.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	path "path/filepath"
	"runtime"
	"strings"
	"time"
)

var usage = `go run internal/tool <COMMAND>

Commands:

- deps       installs development dependencies: stringer, golangci-lint
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
		if err := installDeps(); err != nil {
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

func installDeps() error {
	var goCmd string

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if path, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go (required): %w", err)
	} else {
		goCmd = path

		println("go (required):", goCmd)

		if err := runDep(ctx, goCmd, "version"); err != nil {
			return err
		}
	}

	if stringer, err := exec.LookPath("stringer"); err != nil {
		println("installing stringer")
		println("go install -v golang.org/x/tools/cmd/stringer@latest")

		if err := runDep(ctx, goCmd, "install", "-v", "golang.org/x/tools/cmd/stringer@latest"); err != nil {
			return fmt.Errorf("go install stringer: %w", err)
		}
	} else {
		println("stringer (required):", stringer)
	}

	if linter, err := exec.LookPath("golangci-lint"); err != nil {
		println("installing golangci-lint")

		// ugh
		var installBin string

		if installBinEnv, ok := os.LookupEnv("INSTALLBIN"); ok {
			installBin = installBinEnv
		} else if goBin, ok := os.LookupEnv("GOBIN"); ok {
			installBin = goBin
		} else if goPath, ok := os.LookupEnv("GOPATH"); ok {
			installBin = path.Join(goPath, "bin")
		} else {
			println("golangci-lint: install dir not found. Set INSTALLBIN in your in environment")
			return fmt.Errorf("golangci-lint: unknown install path")
		}

		println("sh", "-c", "\"curl -sSfL "+
			"https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"+
			" | sh -s -- -b '"+installBin+"' v1.55.2\"")

		err = runDep(ctx, "sh", "-c",
			"curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh "+
				"| sh -s -- -b '"+installBin+"' v1.55.2")
		if err != nil {
			return err
		}

		return nil
	} else {
		println("golangci-lint (optional):", linter)
		err = runDep(ctx, linter, "version")
		if err != nil {
			return err
		}
	}

	if docker, err := exec.LookPath("docker"); err != nil {
		println("docker (optional):", err.Error())
	} else {
		println("docker (optional):", docker)
		err = runDep(ctx, docker, "version")
		if err != nil {
			return err
		}
	}

	return nil
}

func runDep(ctx context.Context, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	out, err := c.CombinedOutput()

	println(string(out))

	if err != nil {
		return err
	}

	return nil
}

func dockerBuild() error {
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	//nolint:gosec
	docker := exec.CommandContext(ctx, "docker", "build",
		"-t", "smoynes/elsie",
		"--build-arg", "GOLANG_VERSION="+goVersion,
		"-f", "internal/tool/Dockerfile",
		".",
	)
	out, err := docker.StderrPipe()

	if err != nil {
		return fmt.Errorf("docker: pipe: %w", err)
	}

	if err = docker.Start(); err != nil {
		return fmt.Errorf("docker: build: %w", err)
	}

	println("docker build:")

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
