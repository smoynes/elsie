# <kbd>ELSIE</kbd>: A pedagogical LC-3 CPU simulator #

This project implements a CPU simulator for the LC-3: a little, comprehensive,
and imaginary CPU that was designed as a learning tool for undergraduate
computer-engineering students. It is described in detail in an excellent
textbook, Patt and Patel's *Introduction to Computing Systems: From Bits & Gates
to C/C++ and Beyond*[^1].

The LC-3 instruction set architecture includes:

- a single data type: 16-bit signed words
- word-addressable RAM with 16-bits of address space
- several general purpose registers
- rudimentary arithmetic and logic operations
- memory-mapped I/O
- hard- and software interrupts
- a privileged mode with a basic operating system
- an instruction set that can fit on a single page

It is similar in many respects to the x86 ISA, but is radically simpler in
almost every way. While the LC-3 has 15 opcodes and a few addressing modes, x86
has thousands of instructions, dozens of addressing modes, many data types,
multicore execution, an intricate memory model, many advanced features, and over
40 years of history etched into silicon. Even RISCv8 has hundreds of
instructions and many different addressing modes and extensions. It isn't really
possible anymore for an individual to have a solid understanding of how
contemporary computer hardware works. The LC-3 is totally tractable, though.

A hardware simulator for the LC-3 is a solved problem, of course. The textbook
authors provide a simulator, as well as well as additional learning tools, for
students to use to run and debug their programs. Additionally, there are many
other fantastic tools freely shared online. It has become a bit of a hobby
project among the computing renaissance to build a simulator just for fun or
madness. Your time is almost surely better off spent with any of those.

The fundamental goal of this project is merely to learn about:

- computer architecture
- assembly programming
- operating system fundamentals
- the Age of the PC
- embodying _Shokunin Kishitsu_ (職人気質)[^2]

There remain many problems in computing that are new to me. For the sake of
experiential learning, I plan to read, write, ask questions, find answers, and
solve problems using the old ways and hope that, by holding the master
craftsman's tools, I will gain a better understanding of the past and present of
computing.

I have lots of ideas; Some experiments I might bring into the lab are:

- simulating the LS-3 ISA in software
- building development tools, _e.g._ an assembler, linker, or step-wise debugger
- writing some programs in assembly
- building a compiler for a simple, high-level language[^3]
- extending the ISA, _e.g._ new instructions, data types, math co-processor
- adding new I/O devices, _e.g_ serial console, tape storage, block storage,
  network emulators or adapters
- expanding the operating system, _e.g._ `TRAP` extensions, application services,
  microkernel
- concurrency and parallelism, _e.g._ communicating sequential processes,
  preemptive multitasking, multicore execution
- unleashing <KBD>FACECLOUD</KBD>™️, an ad-supported, smart-contract,
  personal-cloud, privacy-preserving, social-media application based on the LC-3
  💰💰💰

Some of these things seem pretty straightforward, others appear very difficult.
Doing all of them is surely a path ending in madness. You might want to browse
the code herein or follow the project's progress if you enjoy the absurdist
theatrics of a curious software engineer. It might be good trip with
<KBD>ELSIE</KBD> as a guide.

As ever, I am seeking to understand the essence of computing.

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

**IN DEVELOPMENT**: Some instructions are implemented. 🤓

Completed:

- instruction loop

[^1]: https://www.mheducation.com/highered/product/introduction-computing-systems-bits-gates-c-c-beyond-patt-patel/M9781260150537.html
[^2]: _est_, the artisan's spirit
[^3]: Leading to the creation of another dynamically-typed interpreted language -- it is inevitable.
