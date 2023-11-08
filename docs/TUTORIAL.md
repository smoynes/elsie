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
alarmed. You should see towards the end of the output a message that the `Demo
completed`. That is what success looks like:

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

```console
$ elsie demo
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
          PC : 0x3000
          IR : 0x0000 (OP: BR)
         PSR : 0x0007 (N:true Z:true P:true PR:0 PL:0)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3000
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

  - downloaded source code to ELSIE (and, incidentally, the code on which it
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
$ elsie asm COUNTDOWN.asm
```

Well, that isn't satisfying, but no output means success in this case.

```console
$ elsie help asm
Usage:

        elsie asm [-o file.out] file.asm

Assemble source into object code.

Options:
  -debug
    	enable debug logging
  -o filename
    	output filename (default "a.o")
```

```console
$ elsie asm -debug COUNTDOWN.asm
 TIMESTAMP : 2023-10-11T11:51:30-04:00
     LEVEL : DEBUG
    SOURCE : asm.go:66
  FUNCTION : cmd.(*assembler).Run
   MESSAGE : Parsed source
   SYMBOLS : 4
      SIZE : 10
       ERR : <nil>

 TIMESTAMP : 2023-10-11T11:51:30-04:00
     LEVEL : DEBUG
    SOURCE : asm.go:88
  FUNCTION : cmd.(*assembler).Run
   MESSAGE : Writing object
      FILE : a.o

 TIMESTAMP : 2023-10-11T11:51:30-04:00
     LEVEL : DEBUG
    SOURCE : gen.go:56
  FUNCTION : asm.(*Generator).WriteTo
   MESSAGE : Wrote object header
      ORIG : 0x3000

 TIMESTAMP : 2023-10-11T11:51:30-04:00
     LEVEL : DEBUG
    SOURCE : asm.go:103
  FUNCTION : cmd.(*assembler).Run
   MESSAGE : Compiled object
       OUT : a.o
      SIZE : 20
   SYMBOLS : 4
    SYNTAX : 10
```

Ah ha -- we can now see that the assembler has successfully compiled source to
object code.

## Running a program ##

Before running the countdown program, let's take a look at the source code to
get an idea what it is going to do.

```
;;; COUNTDOWN.asm : a program for a tutorial that counts down to blast off
;;; Inputs: R0 is the address of counter.

    .ORIG   x3000               ; Start at the beginning of user-space memory.
    ST      R1,SAVER1           ; Set aside the R1 for the counter value.
    LD      R1,COUNT            ; Load the counter value from the pointer.

    ;;
    ;; Main program loop
    ;;
LOOP:
    BRz     EXIT                ; If counter is zero, exit program.
    ADD     R1,R1,#-1           ; Decrement counter.
    BRnzp   LOOP                ; Loop

    ;;
    ;;  Exit program.
    ;;
EXIT:
    LD      R1,SAVER1
    HALT

    ;;
    ;; Program data
    ;;
COUNT:                          ;
    .FILL   10

SAVER1 .DW x0000
    .END
```

If you have never looked at assembly code, do not be intimidated: what it lacks
in grace and expressiveness it makes up with cold, hard directness. On the one
hand one must communicate to the computer what to do in excruciating detail
without familiar abstractions. On the other, it is all laid bare and nothing is
hidden.

Maybe with the comments and some familiar names in the source you can make a
guess what this program does. It counts down from 10 and then stops. That's it.

You can now try running the program. There is quite a lot of output, arguably
too much, so let is just focus on a single value: the `PC` or _program counter_.
This value points to the next instruction the CPU will execute. Examine its
the value as the program executes:

```console
$ elsie exec countdown.bin 2>&1 | grep 'PC : '
          PC : 0x3000
          PC : 0x3001
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3003
          PC : 0x3004
          PC : 0x3002
          PC : 0x3005
          PC : 0x3006
          PC : 0x1000
          PC : 0x1001
          PC : 0x1002
          PC : 0x1003
          PC : 0x1003
```

The attentive reader will notice that it starts at address `0x3000`, proceeds
take the values `0x3002`, `0x3003`, and `0x3004` in sequence, repeatedly.
Finally, it moves to `0x3005` and `0x3006` before a big jump backwards to
`0x1000` and proceeding to `0x1003`. This little routine, like the steps in a
terrible square dance, are our countdown loop.

<details><summary>Full output…</summary>

If you are curious, take note of the value of R1 throughout the execution
of the program:

```
$ elsie exec -loglevel info exec
 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:17
  FUNCTION : vm.(*LC3).Run
   MESSAGE : START
     STATE :
   !BADKEY :
          PC : 0x3000
          IR : 0x0000 (OP: BR)
         PSR : 0x0300 (N:false Z:false P:false PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0xffff
         MDR : 0x0ff0
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

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3001
          IR : 0x3207 (OP: ST)
         PSR : 0x0300 (N:false Z:false P:false PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3008
         MDR : 0x0000
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

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x2205 (OP: LD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3007
         MDR : 0x000a
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x000a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x000a
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0009
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0009
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0009
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0008
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0008
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0008
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0007
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0007
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0007
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0006
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0006
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0006
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0005
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0005
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0005
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0004
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0004
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0004
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0003
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0003
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0003
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0002
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:11-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0002
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0002
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0001
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0001
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3003
          IR : 0x0402 (OP: BR)
         PSR : 0x0301 (N:false Z:false P:true PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
       INT :
         PL3 : ISR{0xff:Keyboard(status:0x7fff,data:0x2368)}
       REG :
          R0 : 0xffff
          R1 : 0x0001
          R2 : 0xfff0
          R3 : 0xf000
          R4 : 0xff00
          R5 : 0x0f00
          R6 : 0xfe00
          R7 : 0x00f0

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3004
          IR : 0x127f (OP: ADD)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3003
         MDR : 0x127f
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3002
          IR : 0x0ffd (OP: BR)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3004
         MDR : 0x0ffd
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3005
          IR : 0x0402 (OP: BR)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3002
         MDR : 0x0402
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x3006
          IR : 0x2202 (OP: LD)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
         USP : 0xfe00
         SSP : 0x3000
         MCR : 0x8000 (RUN)
         MAR : 0x3008
         MDR : 0x0000
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:105
  FUNCTION : vm.(*LC3).Step
   MESSAGE : instruction raised interrupt
        OP : TRAP: 0x25
       INT : INT: TRAP (0x0000:0x0025)

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x1000
          IR : 0xf025 (OP: TRAP)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x1001
          IR : 0x5020 (OP: AND)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x1002
          IR : 0xe201 (OP: LEA)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:31
  FUNCTION : vm.(*LC3).Run
   MESSAGE : EXEC
     STATE :
   !BADKEY :
          PC : 0x1003
          IR : 0x7040 (OP: STR)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
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

 TIMESTAMP : 2023-10-12T10:12:12-04:00
     LEVEL : INFO
    SOURCE : exec.go:39
  FUNCTION : vm.(*LC3).Run
   MESSAGE : HALTED (HCF)
     STATE :
   !BADKEY :
          PC : 0x1003
          IR : 0x7040 (OP: STR)
         PSR : 0x0302 (N:false Z:true P:false PR:0 PL:3)
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
```

</details>

## Writing a program ##

_Watch this space.__

### Footnotes ###

[^1]: In practice, most machines were rented, I guess. In any case, the ability
to write one's own programs for one's own machine is, perhaps, one of the
essential catalysts of change in the Twentieth-Century post-war period, _IMHO_.

<!-- -*- coding: utf-8-auto -*- -->
