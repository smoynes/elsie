# Tutorial: Building and Running Programs with 𝔼𝕃𝕊𝕀𝔼 #

In this tutorial you will learn how to:

  - install 𝔼𝕃𝕊𝕀𝔼;
  - run a hard-coded demo;
  - execute a simple program;
  - translate a program from <tt>LC3ASM</tt> assembly language to machine
    language; and
  - execute the resulting program.

## Dependencies ##

To install 𝔼𝕃𝕊𝕀𝔼, you will need Go version 1.21, or greater. You can check if
you have a recent enough by running:

```console
$ go version
go version go1.21.4 darwin/amd64
```

If you do not have Go 1.21 installed, you can get it from the Go download page:
<https://go.dev/dl/>. Alternatively, you might be able to use a package manager
for your platform to install a compatible Go version.

## Installation ##

With Go installed, you can now download, build, and install 𝔼𝕃𝕊𝕀𝔼. Run:

```console
$ go install github.com/smoynes/elsie@latset
```

Go will store the program in its `bin` directory. By default, the location is
configured with the `GOBIN` environment variable or, if not set, the default is
the `GOPATH/bin` directory. You can check with:

```console
$ go env GOPATH GOBIN
/home/elsie/go/1.21.4

```

In this case `GOBIN` is unset so the `elsie` command is installed in
`/home/elise/go/1.21.4/bin`. This directory may or may not be present in your
shell's `PATH`. For the sake of the tutorial, we'll assume it is, but do consult
your configuration to add this directory to your system or user configuration.

Say hello:

```console
$ elsie

𝔼𝕃𝕊𝕀𝔼 is a virtual machine and programming tool for the LC-3 educational computer.

Usage:

        elsie <command> [option]... [arg]...

Commands:
        exec     run a program
        asm      assemble source code into object code
        demo     run demo program
        help     display help for commands

Use `elsie help <command>` to get help for a command.
exit status 1

```

## Running the demo ##

𝔼𝕃𝕊𝕀𝔼 includes a silly, hard-coded demo that you can run it with the `demo`
command. You should see a few shocked characters slowly printed and a message of
gratitude. This is what success looks like:

```console
$ elsie demo
⍤⍤
Thank you for demoing!


MACHINE HALTED!

```

The demo does quite a bit of work for little reward. In detail, the demo:

- initializes the virtual machine;
- loads a system image and the program into memory;
- executes the program instructions in sequence according to the control flow;
- outputs a message using BIOS system-calls, themselves small programs;
- halts the virtual machine.

It is not much, but it is an honest program. You can also run the demo with
additional logging enabled. You will see logs for machine startup and its state
after executing each instruction.

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

A hard-coded demo is both impressive andd deeply unsatisfying. It is not enough
to interpret a pointless, pre-written program -- we also want to write our own
pointless programs!

𝔼𝕃𝕊𝕀𝔼 includes a translator that lets us write programs in a simple assembly
dialect called <abbr>LC3ASM</abbr>. We will use the `elsie asm` command to run
the assembler and produce machine code. Later, we will use its output to execute
our program. With this simple command, the full power of the LC-3 is at the tips
of our fingers.

First, save a file named `example.asm`

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

We'll look at the code in detail later on -- for now, we'll run the assembler on
it:

```console
$ elsie asm example.asm
```

No output is produced if successful. You may see error messages if you copied
the code incorrectly or something else went wrong. You can run `elsie asm -log
example.asm` if you would like a chattier assembler.

The output is stored in a file called `a.o`, for lack of a better default. It's
contents should look like:

```
:143000002207240704041081f021127f0ffbf02500050030d9
:00000001ff
```

As you might be able to guess, the machine code is encoded as bytes in something
hexadecimal-y. This is object code, or byte code as it is sometimes called. This
is all that is needed to execute: 𝔼𝕃𝕊𝕀𝔼 loads this data into memory and begins
executing instructions herein.

## Running a program ##

To execute object code for a program, use the `exec` command:

```console
$ elsie exec a.o
54321

MACHINE HALTED!

```

Consider me suitably whelmed.

As with the `asm` command, you can optionally turn on logging when executing
programs. However, in this case logs are sent to a file so the virtual machine
display is not interrupted in the terminal.

```
$ elsie exec -log debug.log a.o
54321

MACHINE HALTED!

$ head debug.log
 TIMESTAMP : 2023-11-30T17:02:14.698750312-05:00
     LEVEL : INFO
    SOURCE : exec.go:138
  FUNCTION : cmd.(*executor).Run.func1
   MESSAGE : Starting machine

 TIMESTAMP : 2023-11-30T17:02:14.69924571-05:00
     LEVEL : INFO
    SOURCE : exec.go:20
  FUNCTION : vm.(*LC3).Run
```


## Execution trace ##

The debug log file will contain a log of machine being initialized and an
execution trace for the program and is invaluable when debugging a program. For
example, consider this log entry:

```
TIMESTAMP : 2023-11-30T17:02:14.699469817-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
          PC : 0x3001
          IR : 0x2207 (OP: LD)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3008
         MDR : 0x0005
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2362
        KBSR : 0x7fff
       REG :
          R0 : 0xffff
          R1 : 0x0005
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2362)}
```

Here we find excruciating detail on the state of the virtual machine after the
first instruction of our program was executed. The most important fields in the
record are:

- **`PC`**: program counter -- a pointer to the next instruction to be
  executed.
- **`IR`**: instruction register -- the value of the previously executed
  instruction and its decoded operation name.
- **`PSR`**: processor status register -- the control flags (or `NZP` statuses),
  priority and privilege levels.
- **`REG`**: registers -- the contents of the general purpose registers.

(The remaining fields in the log record are system-level registers and that
we'll learn about later when we look at system traps and device I/O.)

Given the above, note the values for `PC`, `IR`, `PSR`, and 'R1':

```
 PC : 0x3001
 IR : 0x2207 (OP: LD)
PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
...
 R1 : 0x0005
```

Also, consider the first two lines of our example program:

```asm
      .ORIG 0x3000
      LD   R1,COUNT
```

The very first line is a directive, not an instruction. Directives instruct the
assembler how translate our program, rather than instructions for the virtual
machine. It tells the assembler the address in memory at which the next
instruction is to be found.

The second line of the program is the first instruction that is executed by our
program at address `0x3000` in memory: `LD R1,COUNT`. After execution, the `R1`
register should contain the contents of the memory address labeled `COUNT`.
Indeed! Towards the end of our program we see the line:

```asm
COUNT .DW  5
```
Which, after a bit of hand waving, is a directive that tells the assembler to
store the value `5` in a memory address labeled count. Putting these pieces
together,we see that after executing the first instruction:

- `PC`, the address of the instruction to be executed, is `0x3001`;
- `IR`, the previously executed instruction is `LD`;
- `PSR`, the processor status, indicates the last value loaded was `P`, or
  positive;
- `R1`, the value loaded into the register is `5`.

<!-- -*- coding: utf-8-auto -*- -->
