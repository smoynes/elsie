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

<details>
<summary>Full output…</summary>

```console
$ elsie demo -log
 TIMESTAMP : 2023-11-10T22:12:48.711731-05:00
     LEVEL : INFO
    SOURCE : demo.go:73
  FUNCTION : cmd.demo.Run
   MESSAGE : Initializing machine

 TIMESTAMP : 2023-11-10T22:12:48.712469-05:00
     LEVEL : INFO
    SOURCE : demo.go:84
  FUNCTION : cmd.demo.Run
   MESSAGE : Loading program

 TIMESTAMP : 2023-11-10T22:12:48.712496-05:00
     LEVEL : INFO
    SOURCE : demo.go:128
  FUNCTION : cmd.demo.Run.func4
   MESSAGE : Starting machine

 TIMESTAMP : 2023-11-10T22:12:48.71252-05:00
     LEVEL : INFO
    SOURCE : exec.go:20
  FUNCTION : vm.(*LC3).Run
   MESSAGE : START
     STATE :
   !BADKEY :
          PC : 0x3000
          IR : 0x0000 (OP: BR)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xffff
         MDR : 0x0ff0
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712542-05:00
     LEVEL : INFO
    SOURCE : demo.go:109
  FUNCTION : cmd.demo.Run.func3
   MESSAGE : Starting display

 TIMESTAMP : 2023-11-10T22:12:48.712618-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0420
          IR : 0xf021 (OP: TRAP)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0021
         MDR : 0x0420
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712663-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0421
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0420
         MDR : 0x1dbf
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.7127-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0422
          IR : 0x7380 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfd
         MDR : 0x0000
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712732-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0423
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0422
         MDR : 0x1dbf
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712762-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0424
          IR : 0x7580 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfc
         MDR : 0xfff0
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712805-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0425
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0424
         MDR : 0x1dbf
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712835-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0426
          IR : 0x7780 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfb
         MDR : 0xf000
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712874-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0427
          IR : 0xa210 (OP: LDI)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712913-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0428
          IR : 0x240e (OP: LD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0436
         MDR : 0xbfff
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xbfff
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712947-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0429
          IR : 0x5442 (OP: AND)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0428
         MDR : 0x5442
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.712981-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042a
          IR : 0xb20d (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713022-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042b
          IR : 0xb40c (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713063-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042c
          IR : 0xa60c (OP: LDI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfe04
         MDR : 0x8000
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713099-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042d
          IR : 0x07fd (OP: BR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x042c
         MDR : 0x07fd
         DDR : 0x2368
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713135-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042e
          IR : 0xb00b (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfe06
         MDR : 0x2364
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

⍤ TIMESTAMP : 2023-11-10T22:12:48.713171-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042f
          IR : 0xb208 (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713208-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0430
          IR : 0x6780 (OP: LDR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfb
         MDR : 0xf000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713239-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0431
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0430
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713268-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0432
          IR : 0x6580 (OP: LDR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfc
         MDR : 0xfff0
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.71331-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0433
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0432
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713339-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0434
          IR : 0x6380 (OP: LDR)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfd
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713382-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0435
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0434
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713419-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3001
          IR : 0x8000 (OP: RTI)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdff
         MDR : 0x0300
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713455-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0420
          IR : 0xf021 (OP: TRAP)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0021
         MDR : 0x0420
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713486-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0421
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0420
         MDR : 0x1dbf
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713529-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0422
          IR : 0x7380 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfd
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713563-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0423
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0422
         MDR : 0x1dbf
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713594-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0424
          IR : 0x7580 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfc
         MDR : 0xfff0
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713629-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0425
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0424
         MDR : 0x1dbf
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713661-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0426
          IR : 0x7780 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfb
         MDR : 0xf000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713691-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0427
          IR : 0xa210 (OP: LDI)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713724-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0428
          IR : 0x240e (OP: LD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0436
         MDR : 0xbfff
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xbfff
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713753-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0429
          IR : 0x5442 (OP: AND)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0428
         MDR : 0x5442
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.7138-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042a
          IR : 0xb20d (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713841-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042b
          IR : 0xb40c (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713876-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042c
          IR : 0xa60c (OP: LDI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfe04
         MDR : 0x8000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.713905-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042d
          IR : 0x07fd (OP: BR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x042c
         MDR : 0x07fd
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

⍤ TIMESTAMP : 2023-11-10T22:12:48.793738-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042e
          IR : 0xb00b (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfe06
         MDR : 0x2364
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.793857-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x042f
          IR : 0xb208 (OP: STI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffc
         MDR : 0x0304
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0x8000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.793936-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0430
          IR : 0x6780 (OP: LDR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfb
         MDR : 0xf000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794014-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0431
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0430
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0x0304
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794077-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0432
          IR : 0x6580 (OP: LDR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfc
         MDR : 0xfff0
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794133-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0433
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0432
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0304
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794213-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0434
          IR : 0x6380 (OP: LDR)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfd
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfd
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.79428-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0435
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0434
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794343-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x8000 (OP: RTI)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdff
         MDR : 0x0300
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794421-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0520
          IR : 0xf025 (OP: TRAP)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0025
         MDR : 0x0520
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x2364
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794528-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0521
          IR : 0xe006 (OP: LEA)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0520
         MDR : 0xe006
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794593-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0460
          IR : 0xf022 (OP: TRAP)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0022
         MDR : 0x0460
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794666-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0461
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0460
         MDR : 0x1dbf
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794723-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0462
          IR : 0x7180 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfb
         MDR : 0x236a
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794765-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0463
          IR : 0x1dbf (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0462
         MDR : 0x1dbf
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfa
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794824-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0464
          IR : 0x7380 (OP: STR)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfa
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfa
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794862-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0465
          IR : 0x1220 (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0464
         MDR : 0x1220
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfa
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794905-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0466
          IR : 0x6040 (OP: LDR)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x236a
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfa
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794957-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x046b
          IR : 0x0404 (OP: BR)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0466
         MDR : 0x0404
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfa
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.794993-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x046c
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x046b
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795032-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x046d
          IR : 0x6180 (OP: LDR)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfb
         MDR : 0x236a
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfb
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795076-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x046e
          IR : 0x1da1 (OP: ADD)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x046d
         MDR : 0x1da1
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfc
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795119-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0522
          IR : 0x8000 (OP: RTI)
         PSR : 0x0300 (N:false Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfdfd
         MDR : 0x0300
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x236a
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795163-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0523
          IR : 0xa00a (OP: LDI)
         PSR : 0x0304 (N:true Z:false P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xfffe
         MDR : 0x8000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x8000
          R1 : 0x236a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.79521-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0524
          IR : 0x220a (OP: LD)
         PSR : 0x0301 (N:false Z:false P:true PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x052e
         MDR : 0x7fff
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x8000
          R1 : 0x7fff
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795247-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0525
          IR : 0x5001 (OP: AND)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0524
         MDR : 0x5001
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x7fff
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795289-05:00
     LEVEL : INFO
    SOURCE : exec.go:40
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x0526
          IR : 0xb007 (OP: STI)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x0000 (STOP)
         MAR : 0xfffe
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x7fff
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795335-05:00
     LEVEL : INFO
    SOURCE : exec.go:54
  FUNCTION : vm.(*LC3).Run
   MESSAGE : HALTED (TRAP)
     STATE :
   !BADKEY :
          PC : 0x0526
          IR : 0xb007 (OP: STI)
         PSR : 0x0302 (N:false Z:true P:false PR:System PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x0000 (STOP)
         MAR : 0xfffe
         MDR : 0x0000
         DDR : 0x2364
         DSR : 0x8000
        KBDR : 0x2368
        KBSR : 0x7fff
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x7fff
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-11-10T22:12:48.795387-05:00
     LEVEL : INFO
    SOURCE : demo.go:147
  FUNCTION : cmd.demo.Run
   MESSAGE : Demo completed

```

## Writing a program ##

_Watch this space.__

<!-- -*- coding: utf-8-auto -*- -->
