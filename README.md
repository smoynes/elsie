# `ELSIE`: A pedagogical LC-3 CPU simulator #

This is `ELSIE`, a CPU simulator for the LC-3: a little, comprehensive, and
imaginary CPU that was designed as a learning tool for undergraduate
computer-engineering students. It is described in detail in an excellent
textbook, Patt and Patel's *Introduction to Computing Systems: From Bits & Gates
to C/C++ and Beyond*[^1].

The LC-3 instruction set architecture includes:

- a single data type: signed integers stored as 16-bit words
- word-addressable RAM with 16-bit  address space
- several general purpose registers
- rudimentary arithmetic and logic operations
- memory-mapped I/O
- hard- and software interrupts
- a privileged mode with a basic operating system
- an instruction set compact enough to fit on a single page

While similar in many respects to the x86 ISA, the LC-3 is radically simpler in
almost every way. Unlike the sprawling x86, with thousands of instructions,
dozens of addressing modes, multicore execution, an intricate memory model, many
advanced features, and over 40 years of history etched into silicon, the LC-3
remains a tractable system that is comprehensible by an individual.

Hardware simulators already exist for the LC-3 architecture. The textbook
publishers provide one and there are many others freely available online. This
one is a reinvention of the wheel with the aim of learning about:

- computer architecture
- assembly programming
- operating systems
- personal computing history
- _Shokunin Kishitsu_ (職人気質), the artisan's spirit

There remain many computer things that baffle me. Through experiential learning,
I will read, write, ask questions, find answers, and solve problems. It is to be
hoped that by building something esoteric using the old ways, by holding the
master craftsman's tools in my hands, I will gain a better understanding of
computing's past and present.

I have lots ideas for experiments to bring into the lab. I've thought about:

- simulating the LS-3 ISA in software
- hand-coding machine code from instructions
- writing some programs in assembly
- running some programs written by others
- building development tools like an assembler, linker, loader, and step-wise
  debugger
- building a simple compiler for high-level language[^2]
- extending the ISA with new instructions, data types, or a math co-processor
- adding new I/O devices like a serial console, a tape storage, block storage,
  or network emulators and adapters
- expanding the operating system, with `TRAP` extensions, runtime library,
  application services, or even a microkernel
- concurrency and parallelism, _e.g._ communicating sequential processes,
  preemptive multitasking, multicore execution

Admittedly, some of these experiments seem pretty straightforward, while others
appear daunting and complex; some are immediate goals but most mere thought
experiments. Trying to do all of that work might indeed be a path in madness.

Maybe, with some dedication and blind ambition I will, finally realize my
lifelong ambition: to develop FACECLOUD™️, an ad-supported, privacy-preserving,
social-media smart-contract for personal clouds based, of course, on the LC-3
ISA! 💰💰💰

In the meantime, feel free to browse the code or follow the project if you enjoy
the absurdist theatrics of a curious software engineer. As ever, I seek to
understand the essence of computing.

You are welcome to reach out if you'd like to join me on this journey of
exploration.

----

<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">
    <img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by-nc-sa/4.0/88x31.png" />
</a>
<br />

This work is licensed under a
<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">
Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License
</a>.

----

## Status ##

**IN DEVELOPMENT**

Focus right now:

- memory mapped i/o

Up next:

- ISRs
- I/O interrupts

Completed:

- instruction loop
- operations:
  - BR
  - NOT
  - AND
  - ADD
  - LD
  - LDI
  - LDR
  - LEA
  - ST
  - STI
  - STR
  - JMP/RET
  - JSR
  - JSRR
  - TRAP
  - RTI
  - RESV
- exceptions
- traps
- memory mapped I/O

----

## Dedication ##

This work is dedicated to the MCM/70[^3] and its pioneering designers.

----

[^1]: https://www.mheducation.com/highered/product/introduction-computing-systems-bits-gates-c-c-beyond-patt-patel/M9781260150537.html
[^2]: Leading to the creation of another dynamically-typed, interpreted language -- it is inevitable.
[^3]: https://en.wikipedia.org/wiki/MCM/70
