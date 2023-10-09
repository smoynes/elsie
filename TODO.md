# TODO #

- ASM:
  - [.] document grammar more completely
  - directives remaining:
    - [ ] .END
    - [ ] .EXTERNAL
- [ ] LOAD: object loader
- DOCS:
  - [ ] tutorial
  - [ ] design
- [ ] MONITOR: trap, exception, and interrupt routines
- [ ] LINK: code linker
- [ ] DUMP: hex encoder
- [ ] CLI: sub commands for vm, tools, terminal, shell
- [ ] LOG: program output to STDOUT, logging output to STDERR (unless in
      demo)
- [ ] TERM: finish the terminal I/O
- [ ] REPL: step debugger shell
- [ ] TIMER: simple timer device
- [ ] NEWLANG: interpreted language; threaded compiler

... more to come.

## DID ##

- ASM:
  - [x] code generation
  - [x] completed directives:
    - [x] .STRINGZ: strangz
    - [x] .BLKW: block words
    - [x] .DW: define word
    - [x] .FILL: fill word
    - [x] .ORIG: origin
  - [x] cli command
  - [x] memory layout
  - [x] cleaner error handling
  - [x] simple parser: regexp
  - [x] symbol table
- VM:
  - [x] instruction loop
  - [x] instructions
  - [x] service exceptions
  - [x] service traps
  - [x] memory mapped I/O
  - [x] hardware interrupts
