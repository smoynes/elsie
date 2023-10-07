# TODO #

- [ ] DOCS:
  - [ ] tutorial
  - [ ] design
- [ ] ASM:
  - [x] simple parser: regexp
  - [x] symbol table
  - [.] code generation: ~5/16
  - [.] directives
    - [x] .BLKW: block words
    - [x] .DW: define word
    - [x] .FILL: fill word
    - [x] .ORIG: origin
    - [ ] .STRINGZ: strangz
    - [ ] .END
  - [.] memory layout
  - [.] cli command
  - [x] cleaner error handling
  - [ ] document grammar
- [ ] LOAD: object loader
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
