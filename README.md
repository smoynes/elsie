# elsie: A pedagogical LC-3 CPU simulator #

This project implements a hardware emulator for the LC-3: a little computer with
a simple ISA from Patt and Patel's
[*Introduction to Computing Systems: from bits & gates to C/C++ and beyond*](https://www.mheducation.com/highered/product/introduction-computing-systems-bits-gates-c-c-beyond-patt-patel/M9781260150537.html).

The LC-3 is a 16-bit architecture with word-addressable RAM with 16-bits of
address space, several general purpose registers, rudimentary arithmetic and
logic instructions and simple serial I/O. It has a privileged mode that is used
by a basic operating system. It is similar in many respects to the x86 ISA, but
is radically simpler in almost any way. While the LC-3 has 15 instructions, x86
has hundreds.

A hardware simulator for the LC-3 is a solved problem, of course. The textbook
authors provide a simulator, as well as well as additional learning tools, for
students to use to run and debug their programs. There are many other fantastic
simulators freely shared online. Your time is better off spent with any of
those.

The goal of this project is merely to learn about:

- computer architecture;
- simulating hardware in software;
- assembly programming;
- building a simple assembler and linker;
- operating system fundamentals.

As ever, I seek to understand the essence of computing.

<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by-nc-sa/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License</a>.
