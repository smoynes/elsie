# Tutorial: Building and Running Programs with 𝔼𝕃𝕊𝕀𝔼 #

In this tutorial you will learn how to:

  - install 𝔼𝕃𝕊𝕀𝔼;
  - run a hard-coded demo;
  - execute a simple program;
  - translate a program from <tt>LC3ASM</tt> assembly language to machine
    language; and
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

While a hard-coded demo is impressive, it is also deeply unsatisfying. It is not
enough to interpret a pointless, pre-written program -- we also want to write
our own pointless programs for our machine to interpret.

𝔼𝕃𝕊𝕀𝔼 includes a translator that lets us write programs in a simple assembly
dialect called <tt>LC3ASM</tt>. We will use the `elsie asm` command to run the
assembler and produce machine code. Later, we will execute the program.

First save a file named `example.asm`

```asm
      .ORIG x3000
      LD   R1,COUNT
      LD   R2,ASCII
LOOP  BRz  EXIT
      ADD  R0,R2,R1
      TRAP x21
      ADD  R1,R1,#-1
      BR   LOOP
EXIT  HALT
COUNT .DW  5
ASCII .DW  48
      .END
```

We'll look at the code in detail later -- for now, we'll run the assembler on
it:

```console
$ elsie asm example.asm
```

No output is produced if successful. You may see error messages if you copied
the code incorrectly. The output is stored in a file called `a.o`, for lack of a
better default. It's contents should look like:

```
:143000002207240704041081f021127f0ffbf02500050030d9
:00000001ff
```

As you might be able to guess, the machine code is encoded as bytes in something
hexadecimal-y. This is all that is needed to be executed.

```console
$ elsie exec a.o
54321

MACHINE HALTED!

```

Consider me suitably whelmed.

<!-- -*- coding: utf-8-auto -*- -->
