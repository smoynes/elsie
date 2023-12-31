================================================================================

                       𝔼𝕃𝕊𝕀𝔼: A Pedagogical LC-3 Emulator

================================================================================

            The path is made in the walking of it. -- Zhuangzi

This is 𝔼𝕃𝕊𝕀𝔼, an exploration of the LC-3: a model computer architecture that is
simple, small, and imaginary.

The project includes:

 - a virtual machine that executes instructions;
 - an assembler for translating LC3ASM source to machine code;
 - a loader that puts programs into memory;
 - a system monitor that implements system calls;
 - virtual devices for display and keyboard I/O;
 - many unnecessary words by your author; and
 - maybe, more to come…

It's purpose is as a tool for learning and a path of discovery.

-----------------------------
         Background
-----------------------------

The  LC-3  computer  architecture  was   designed  as  a  educational  tool  for
undergraduate computer-engineering  students. It  is described  in detail  in an
excellent textbook,  _Introduction to  Computing Systems: From  Bits &  Gates to
C/C++ and Beyond_ (3/e), by Yale Patt and Sanjay Patel.

The LC-3 instruction set and architecture (ISA) includes:

  - a single data type: signed 16-bit words;
  - word-addressable RAM with 16-bit address space;
  - several general purpose registers;
  - three rudimentary arithmetic and logic operations;
  - memory-mapped I/O;
  - hard- and software interrupts; and
  - an instruction set compact enough to fit on a single page.

Far  from an  abstract machine,  the text  begins with  transistors and  digital
logic. From there,  it builds upon the  titular bits and gates  and describes an
entire computer architecture in detail including the control-unit state-machine,
data  and  I/O paths.  Upon  this  computer,  assembly,  C and  C++  programming
languages  are described.  It  is fascinating.  As  far as  I  know, a  complete
hardware implementation has never been physically  built but, I can imagine, the
text  will be  invaluable when  humanity has  to recreate  computers from  first
principles.

While similar  in many respects to  more familiar x86  or ARM ISAs, the  LC-3 is
radically simpler in almost every way.  Unlike the sprawling x86, with thousands
of instructions, dozens  of addressing modes, multicore  execution, an intricate
memory model, and over 40 years of history etched into silicon, the LC-3 remains
a tractable system that is comprehensible by an individual.

It is  a lot closer to  a PDP/7 machine than  anything you have in  your home or
pocket. Nevertheless,  it still takes quite  a lot of effort  to understand well
enough to write programs. Our effort pays dividends in knowledge.

-----------------------------
       Project Goals
-----------------------------

𝔼𝕃𝕊𝕀𝔼 is not useful: it isn't complete and doesn't work well. In those terms, it
is not good software. However, to the  author, the project is not intended to be
useful to others, not in utilitarian terms, at least.

Neither is 𝔼𝕃𝕊𝕀𝔼  novel: simulators already exist for the  LC-3 architecture, of
course. The  textbook publishers provide  one and  there are many  others freely
available online. This one is admittedly a mere reinvention of the wheel.

That said, the  gift the design and  engineering process affords is  that it can
reveal  something  about  either  our world  or  ourselves.  Hopefully  building
something useless and retreading well worn paths will expose something essential
and fundamental.

Personally,  there remain  many  Computer  Things that  baffle  me. Despite  ten
thousand hours of computing,  I still feel lost when it  comes to the rudiments:
computer architecture, assembly programming, and operating systems. This project
is an artifact  of my learning process  and a trace of my  explorations into the
field.

It is  to be hoped  that by trying  my hand at the  old methods, by  holding the
craftsman's  tools and  building something  unnecessary,  I will  gain a  better
understanding of the essence of computing.  And yet, if nothing else is achieved
than reading a few  books, learning some ideas, and hearing  a few good stories,
it will have been worth it in the end.

-----------------------------
       Get in Touch
-----------------------------

I have  lots to learn, many  ideas for experiments to  run, and even a  few more
project plans to bring to my workbench.

You are welcome to reach out if:

  - you're a fellow learner;
  - if you find this project useful or buggy;
  - if you have any ideas or questions or feedback; or,
  - if you have a story to share, especially.

You     can     start     a     discussion     on     the     GitHub     project
<https://github.com/smoynes/elsie/discussions>  or  you're   always  welcome  to
contact me directly through my GitHub profile.

Please follow  the project  if you  enjoy the absurdist  theatrics of  a curious
software engineer. As ever, I simply seek to understand the essence of computing
and to embody _Shokunin Kishitsu_ (職人気質), the artisan's spirit.

-----------------------------
        Documentation
-----------------------------

- README.txt           You are here.
- TUTORIAL.md          A trailhead for users.
- DEVGUIDE.txt         Development guidebook.
- RESOURCE.txt         Inspirations and references.
- LICENCE.txt          Terms of use.
- CODE_OF_CONDUCT.txt  Behave yourself.

-----------------------------
         Dedication
-----------------------------

This work is dedicated to the MCM/70 and its pioneering designers.
<https://en.wikipedia.org/wiki/MCM/70>

-----------------------------

𝔼𝕃𝕊𝕀𝔼 © 2023 by Scott Moynes is licensed under CC BY-SA 4.0. See LICENCE.txt for
terms. Send your lawyers here:
<https://creativecommons.org/licenses/by-sa/4.0/?ref=chooser-v1>
