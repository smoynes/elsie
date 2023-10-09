# TODO #

- [ ] DOCS:
  - [ ] tutorial
  - [ ] design
- [.] ASM:
  - [.] code generation: ~13/16
  - [.] directives
    - [ ] .END
    - [ ] .EXTERNAL
    - [x] .STRINGZ: strangz
    - [x] .BLKW: block words
    - [x] .DW: define word
    - [x] .FILL: fill word
    - [x] .ORIG: origin
  - [x] cli command
  - [.] document grammar more completely
  - [x] memory layout
  - [x] cleaner error handling
  - [x] simple parser: regexp
  - [x] symbol table
- [ ] LOAD: object loader
- [ ] LINK: code linker
- [ ] DUMP: hex encoder
- [ ] CLI: sub commands for vm, tools, terminal, shell
- [ ] LOG: program output to STDOUT, logging output to STDERR (unless in
      demo)
- [ ] TERM: finish the terminal I/O
- [ ] MONITOR: trap, exception, and interrupt routines
- [ ] REPL: step debugger shell
- [ ] TIMER: simple timer device
- [ ] NEWLANG: interpreted language; threaded compiler

... more to come.

## DID ##

- [x] instruction loop
- [x] instructions
- [x] service exceptions
- [x] service traps
- [x] memory mapped I/O
- [x] hardware interrupts
