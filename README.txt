                       ELSIE: A Pedagogical LC-3 Emulator
================================================================================

            The path is made in the walking of it. -- Zhuangzi

This is ELSIE, an exploration of the LC-3: a little computer that is simple,
comprehensive, and imaginary.

The project includes:

 - a virtual machine, including CPU and I/O devices;
 - an assembler, for producing machine code from LCASM assembly language;
 - a lot of unnecessary words from your author; and
 - hopefully, more to come.

As a technical project, ELSIE is not very useful: it is more like a story, a
performance. As such, it is much more than mere software.

    Background
------------------

The LC-3 computer was designed as a educational tool for undergraduate
computer-engineering students and is described in detail in an excellent
textbook, Patt and Patel's _Introduction to Computing Systems: From Bits & Gates
to C/C++ and Beyond_[^1], 3ed.

The LC-3 instruction set and architecture includes:

  - a single data type: signed, two's-complement  16-bit words;
  - word-addressable RAM with 16-bit address space;
  - several general purpose registers;;
  - rudimentary arithmetic and logic operations;
  - memory-mapped I/O;
  - hard- and software interrupts; and
  - an instruction set compact enough to fit on a single page.

Far from an abstract machine, the text begins with transistors and digital
logic. From there, it builds upon the titular bits and bytes and describes in
detail the entire computer architecture including the control-unit
state-machine, data and I/O paths. It is fascinating. As far as I know, a
complete implementation has never been physically built but, I can imagine, the
text will be invaluable when humanity has to recreate computers from first
principles.

While similar in many respects to the x86 or ARM ISAs, the LC-3 is radically
simpler in almost every way. Unlike the sprawling x86, with thousands of
instructions, dozens of addressing modes, multicore execution, an intricate
memory model, and over 40 years of history etched into silicon, the LC-3 remains
a tractable system that is comprehensible by an individual.

It is a lot closer to a PDP/11 machine than anything you have in your home or
pocket. Nevertheless, it is still takes quite a lot of effort to understand well
enough to write programs.

    Project Goals
---------------------

ELSIE is not novel: hardware simulators already exist for the LC-3 architecture,
of course. The textbook publishers provide one and there are many others freely
available online.[^4] This one is admittedly a mere reinvention of the wheel.
That said, the gift the design and engineering process affords is that sometimes
it reveals something fundamental about either our world or ourselves. So, I
think, it is worth retreading the path.

Personally, there remain many Computer Things that baffle me. Despite ten
thousand hours of computing, I still feel lost when it comes to some of the
rudiments of the field:

  - computer architecture;
  - assembly programming;
  - operating systems;
  - computing history.

I had a thought that a good way to learn about these topics was to get my boots
dirty and learn the basics by building simple things. This project is an
artifact of my process.

It is to be hoped that by trying my hand at the old methods, by holding the
master craftsman's tools, by building something both cute and useless, I will
gain a better understanding of the essence of computing. If nothing else is
achieved than learning a bit, exploring some ideas, and hearing a few good
stories, it will have been worth it.

I have lots to learn, many ideas for experiments, and a few more plans to bring
to my workbench. I've thought about:

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

    Get in Touch
--------------------

You are welcome to reach out if:

  - you're a fellow learner;
  - if you find this project useful (or buggy); or,
  - if you have any questions or feedback.

You can start a [discussion](https://github.com/smoynes/elsie/discussions) on
this project or you're welcome to contact me directly through my GitHub profile.

Please feel free to browse the code or follow the project if you enjoy the
absurdist theatrics of a curious software engineer. As ever, I simply seek to
understand the essence of computing and to embody _Shokunin Kishitsu_ (職人気質),
the artisan's spirit.

    Dedication
------------------
This work is dedicated to the MCM/70[^3] and its pioneering designers.

    Footnotes
-----------------

[^1]: https://www.mheducation.com/highered/product/introduction-computing-systems-bits-gates-c-c-beyond-patt-patel/M9781260150537.html
[^2]: Leading to the creation of another dynamically-typed, interpreted language -- it is inevitable.
[^3]: https://en.wikipedia.org/wiki/MCM/70
[^4]: You can find references to some other tools and some useful resources in
      RESOURCES.txt

-----------------

ELSIE © 2023 by Scott Moynes is licensed under CC BY-SA 4.0.
See LICENCE.txt for terms. Send your lawyers here:
https://creativecommons.org/licenses/by-sa/4.0/?ref=chooser-v1
