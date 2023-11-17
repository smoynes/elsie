# Tutorial: Building and Running Programs with 𝔼𝕃𝕊𝕀𝔼 #

In this tutorial you will learn how to:

  - install 𝔼𝕃𝕊𝕀𝔼;
  - run a hard-coded demo.
  - execute a simple program;
  - translate a program from LC3ASM assembly language to machine language; and
  - execute the resulting program.

## Installation ##

To install the 𝔼𝕃𝕊𝕀𝔼, you will need Go version 1.21, or greater. To check if you
have a good version, run:

```console
$ go version
go version go1.21.4 darwin/amd64
```

If you do not have Go 1.21 installed, you can get it from the Go download page:
https://go.dev/dl/

Alternatively, you might be able to use a package manager for your platform to
install a compatible Go version.

With Go installed, you can now download, build, and install 𝔼𝕃𝕊𝕀𝔼. Run:

```console
$ go install github.com/smoynes/elsie@latset
```

Go will store the program in its `bin` directory. By default, the location is
configured with the `GOBIN` environment variable or the `GOPATH/bin` directory.
You can check with:

```console
$ go env GOPATH GOBIN
/home/elsie/go/1.21.4

```

In this case `GOBIN` is unset so the `elsie` command is installed in
`/home/elise/bin/1.21.4/bin`. This directory may or may not be present in your
shell's `PATH`. For the sake of the tutorial, we'll assume it, is but do consult
your configuration to add this directory to your system or user configuration.

Say hello:

```console
$ elsie

𝔼𝕃𝕊𝕀𝔼 is a virtual machine and programming tool for the LC-3 educational computer.

Usage:

        elsie <command> [option]... [arg]...

Commands:
  exec                 run a program
  asm                  assemble source code into object code
  demo                 run demo program
  help                 display help for commands

Use `elsie help <command>` to get help for a command.
exit status 1

```

## Running the demo ##

𝔼𝕃𝕊𝕀𝔼 includes a silly, hard-coded demo that you can run it with the `demo`
command. You should see a few shocked characters printed and a message of
gratitude. This is what success looks like:

```console
$ elsie demo
⍤⍤
Thank you for demoing!


MACHINE HALTED!

```

The demo initialize the machine, outputs a message using BIOS system-calls, and
halts the machine. It is not much, but it is an honest program.

You can also run the demo with additional logging enabled. You will see logs for
machine startup and its state after executing each instruction.

```console
$ elsie demo -log
 TIMESTAMP : 2023-11-10T22:11:56.720592-05:00
     LEVEL : INFO
    SOURCE : demo.go:73
  FUNCTION : cmd.demo.Run
   MESSAGE : Initializing machine

 TIMESTAMP : 2023-11-10T22:11:56.721381-05:00
     LEVEL : INFO
    SOURCE : demo.go:84
  FUNCTION : cmd.demo.Run
...
 TIMESTAMP : 2023-11-10T22:12:10.90828-05:00
     LEVEL : INFO
    SOURCE : demo.go:147
  FUNCTION : cmd.demo.Run
   MESSAGE : Demo completed
```

## Writing a program ##

_Watch this space.__

<!-- -*- coding: utf-8-auto -*- -->
