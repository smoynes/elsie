# Tutorial: Building and Running Programs #

In this tutorial you will learn more about:

  - installing ELSIE;
  - writing simple machine code programs and executing them;
  - building a few more complicated programs in assembly language.

## Installing ELSIE ##

- You will need Go version 1.21, or greater, installed.
- Visit the Go download page:  https://go.dev/dl/
- Install a version greater than 1.21.

You can check if you have a good version, run:

```console
$ go version
go version go1.21.1 darwin/amd64
```

Finally, you can now download, build, and install ELSIE. Run:

```console
$ go install github.com/smoynes/elsie
```

Say hello:

```console
$ elsie

ELSIE is a virtual machine for the LC-3 educational computer.

Usage:

        elsie <command> [option]... [arg]...

Commands:
  demo                 run demo program
  asm                  assemble source code into object code
  help                 display help for commands

Use `elsie help <command>` to get help for a command.

```

## Running the demo ##

ELSIE includes a silly demo. Run it:

```console
$ elsie demo
```

You should see a bunch of debug output spammed to your terminal. Do not be
alarmed. You should see, towards the end of the output a message that the `Demo
completed`. That is what success looks like.

```console
 TIMESTAMP : 2023-10-10T23:00:50-04:00
     LEVEL : INFO
    SOURCE : exec.go:39
  FUNCTION : vm.(*LC3).Run
   MESSAGE : HALTED (HCF)
     STATE :
        VM :
          PC : 0x1003
          IR : 0x7040 (OP: STR)
         PSR : 0x0002 (N:false Z:true P:false PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x0000 (STOP)
         MAR : 0xfffe
         MDR : 0x0000
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2362)}
       REG :
          R0 : 0x0000
          R1 : 0xfffe
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

TIMESTAMP : 2023-10-10T23:00:50-04:00
     LEVEL : INFO
    SOURCE : demo.go:132
  FUNCTION : cmd.demo.Run
   MESSAGE : Demo completed
```

<Details>
<summary>Full output…</summary>

```
 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : demo.go:57
  FUNCTION : cmd.demo.Run
   MESSAGE : Initializing machine

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : demo.go:60
  FUNCTION : cmd.demo.Run
   MESSAGE : Loading trap handlers

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : demo.go:113
  FUNCTION : cmd.demo.Run
   MESSAGE : Loading program

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : demo.go:125
  FUNCTION : cmd.demo.Run
   MESSAGE : Starting machine

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : exec.go:17
  FUNCTION : vm.(*LC3).Run
   MESSAGE : START
     STATE :
        VM :
          PC : 0x0300
          IR : 0x0000 (OP: BR)
         PSR : 0x0007 (N:true Z:true P:true PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0300
         MDR : 0xf025
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : ERROR
    SOURCE : exec.go:105
  FUNCTION : vm.(*LC3).Step
   MESSAGE : instruction raised interrupt
        OP : TRAP: 0x25
       INT : INT: TRAP (0x0000:0x0025)
    HANDLE : INT: TRAP (0x0000:0x0025)

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
        VM :
          PC : 0x1000
          IR : 0xf025 (OP: TRAP)
         PSR : 0x0007 (N:true Z:true P:true PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x0025
         MDR : 0x1000
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
        VM :
          PC : 0x1001
          IR : 0x5020 (OP: AND)
         PSR : 0x0002 (N:false Z:true P:false PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x1000
         MDR : 0x5020
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0x0000
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
        VM :
          PC : 0x1002
          IR : 0xe201 (OP: LEA)
         PSR : 0x0002 (N:false Z:true P:false PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x1003
         MDR : 0xfffe
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0xfffe
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
        VM :
          PC : 0x1003
          IR : 0x7040 (OP: STR)
         PSR : 0x0002 (N:false Z:true P:false PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x0000 (STOP)
         MAR : 0xfffe
         MDR : 0x0000
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0xfffe
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : exec.go:39
  FUNCTION : vm.(*LC3).Run
   MESSAGE : HALTED (HCF)
     STATE :
        VM :
          PC : 0x1003
          IR : 0x7040 (OP: STR)
         PSR : 0x0002 (N:false Z:true P:false PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x0000 (STOP)
         MAR : 0xfffe
         MDR : 0x0000
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0x0000
          R1 : 0xfffe
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfdfe
          R7 : 0x00f0

 TIMESTAMP : 2023-10-11T10:25:48-04:00
     LEVEL : INFO
    SOURCE : demo.go:132
  FUNCTION : cmd.demo.Run
   MESSAGE : Demo completed
```

</code>
</details>

The demo does not do much and is quite noisy about it. It loads a `HALT`
instruction into program memory, initializes a trap service routine that stops
the VM, and then runs the program.

Notice, in particular, the value for `PC` when the machine halted: `0x1003`.
This means the next instruction to be executed, had the machine not stopped, is
held at address `0x1003`. This is notable because our program started execution
at a different address, `0x3000`: it seems our machine did a bit of work to stop
doing any more.

## Assemble a program ##

You, my dear reader, have already accomplished so much. If you have reached this
point in our tutorian journey, you have:

  - downloaded source code to ELSIE (and, incidentally, the code one which it
    depends);
  - compiled the code to create a virtual computer;
  - had the machine execute a hard-coded program.

Consider, none of that would have been possible, not even imaginable, one
hundred years ago. Only a few researchers, defence organizations, and spies had
this capability eighty years ago. A decade later, perhaps, large commercial
organizations started writing and running their own programs on their own
machines[^1].

Yet, it is not really quite what we imagined computing to be. Next, you will
create machine code from source. In this directory you will find an assembly
program. To translate the source to an object file containing machine code, run:

```console
$ elsie asm countdown.asm
```

### Footnotes ###

[^1]: In practice, most machines were rented, I guess. In any case, the ability
to write one's own programs for one's own machine is, perhaps, one of the
essential catalysts of change in the Twentieth-century post-war period, _IMHO_.

<!-- -*- coding: utf-8-auto -*- -->

