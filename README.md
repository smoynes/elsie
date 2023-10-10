# `ELSIE`: A pedagogical LC-3 emulator #

> The path is made in the walking of it. -- Zhuangzi

This is `ELSIE`, a virtual machine for the LC-3: a little computer that is
simple, comprehensive, and imaginary.

The computer was designed as a learning tool for undergraduate
computer-engineering students and is described in detail in an excellent
textbook, Patt and Patel's *Introduction to Computing Systems: From Bits & Gates
to C/C++ and Beyond*[^1], 3ed.

## LC-3 Background ##

The LC-3 instruction set architecture includes:

  - a single data type: signed integers stored as 16-bit words;
  - word-addressable RAM with 16-bit address space;
  - several general purpose registers;;
  - rudimentary arithmetic and logic operations;
  - memory-mapped I/O;
  - hard- and software interrupts; and
  - an instruction set compact enough to fit on a single page.

Far from an abstract machine, the text begins with transistors and digital logic
and describes in detail the entire computer architecture including the
control-unit state-machine, data and I/O paths. It is fascinating.

While similar in many respects to the x86 or ARM ISAs, the LC-3 is radically
simpler in almost every way. Unlike the sprawling x86, with thousands of
instructions, dozens of addressing modes, multicore execution, an intricate
memory model, and over 40 years of history etched into silicon, the LC-3 remains
a tractable system that is comprehensible by an individual. It is a lot closer
to PDP/11 machines than anything you have in your home or pocket.

## Project Goals ##

Personally, there remain many computer things that baffle me. Despite ten
thousand hours of computing, I feel lost and uncertain when it comes to some of
the rudiments of the field:

  - computer architecture;
  - assembly programming;
  - operating systems;
  - computing history.

This project is not novel: hardware simulators already exist for the LC-3
architecture, of course. The textbook publishers provide one and there are many
others freely available online.[^4] This one is admittedly a mere reinvention of
the wheel. Nevertheless, the design and engineering process sometimes reveals
something fundamental about either ourselves or our world, so it is worth
retreading the path.

This slightly absurd project's purpose is to explore these topics through
experiential learning: gaining a deeper understanding of how a thing is done by
simply doing the thing. It is to be hoped that by trying the old methods, by
holding the master craftsman's tools, by building something both cute and
useless, I will gain a better understanding of the essence of computing. If
nothing else is achieved than learning a bit, exploring some ideas, and hearing
a few good stories, it will have been worth it.

I have lots ideas for experiments and projects to bring to my workbench. I've
thought about:

  - simulating the LS-3 ISA in software;
  - building development tools like an assembler, linker, loader, and step-wise
    debugger;
  - writing some programs in assembly, translating to machine code;
  - running some programs written by others;
  - building a simple compiler for high-level language[^2];
  - extending the ISA with new instructions, data types, or a math co-processor;
  - adding new I/O devices tape storage, block storage, or network emulators and
    adapters;
  - expanding the operating system, with new system calls, a runtime library,
    IPC services, or even a microkernel;
  - concurrency and parallelism, _e.g._ co-operative sequential processes,
    preemptive multitasking, multicore execution.

Admittedly, some of these experiments are pretty straightforward, while others
appear daunting and complex; some are immediate goals, but most are mere thought
experiments. Trying to do all of that work might be a path in madness.

## Get in Touch ##

You are welcome to reach out if you're a fellow learner, if you find this
project useful (or buggy), or you have any questions or feedback. You can start
a [discussion](https://github.com/smoynes/elsie/discussions) on this project or
contact me directly through my GitHub profile.

Please feel free to browse the code or follow the project if you enjoy the
absurdist theatrics of a curious software engineer. As ever, I simply seek to
understand the essence of computing and to embody _Shokunin Kishitsu_ (職人気質),
the artisan's spirit.

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

<details>
<summary>
<strong>IN DEVELOPMENT</strong>
</summary>

Focus right now:

ASM: assembler
  - generates code for a few opcodes
  - compatible with textbook
  - not enough error handling
  - missing some opcodes

On deck:

- BIOS: interrupt routines

See [TODO.md](`TODO.md`) for more ideas.

</details>

----

## Dedication ##

This work is dedicated to the MCM/70[^3] and its pioneering designers.

----

## Footnotes ##

[^1]: https://www.mheducation.com/highered/product/introduction-computing-systems-bits-gates-c-c-beyond-patt-patel/M9781260150537.html
[^2]: Leading to the creation of another dynamically-typed, interpreted language -- it is inevitable.
[^3]: https://en.wikipedia.org/wiki/MCM/70
[^4]: You can find references to some other tools and some useful resources in
  [./RESOURCES.txt](`RESOURCES.txt`)
