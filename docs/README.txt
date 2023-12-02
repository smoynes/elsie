                       𝔼𝕃𝕊𝕀𝔼: A Pedagogical LC-3 Emulator
================================================================================

            The path is made in the walking of it. -- Zhuangzi

This is 𝔼𝕃𝕊𝕀𝔼, an exploration of the  LC-3: a model computer that is small, simple, and
imaginary.

The project includes:

 - a virtual machine that executes instructions;
 - an assembler for translating LC3ASM source to machine code;
 - a loader that puts programs into memory;
 - a system monitor that implements system calls;
 - virtual devices  for display and keyboard I/O;
 - many unnecessary words by your author; and
 - maybe, more to come…

As a technical project, 𝔼𝕃𝕊𝕀𝔼 is not  useful: it isn't complete and doesn't work
well. In  those terms,  it is not  good software. However,  for the  author, the
project is  not really  intended to  be useful to  others in  utilitarian terms.
Rather,  it  is  meant  to  as  an exercise  in  learning  more  about  computer
architecture,  systems programming,  and oneself.  As such,  it is  more like  a
story, a performance, or a trail through the woods and is essential.

-----------------------------
         Background
-----------------------------

The  LC-3  computer  was  designed  as  a  educational  tool  for  undergraduate
computer-engineering  students  and  is  described in  detail  in  an  excellent
textbook, _Introduction  to Computing Systems:  From Bits  & Gates to  C/C++ and
Beyond_ (3/e), by Yale Patt and Sanjay Patel.

The LC-3 instruction set and architecture includes:

  - a single data type: signed 16-bit words;
  - word-addressable RAM with 16-bit address space;
  - several general purpose registers;
  - three rudimentary arithmetic and logic operations;
  - memory-mapped I/O;
  - hard- and software interrupts; and
  - an instruction set compact enough to fit on a single page.

Far  from an  abstract machine,  the text  begins with  transistors and  digital
logic. From there,  it builds upon the  titular bits and bytes  and describes an
entire computer architecture in detail including the control-unit state-machine,
data  and  I/O  paths.  It  is  fascinating.  As  far  as  I  know,  a  complete
implementation has never been physically built but, I can imagine, the text will
be invaluable when humanity has to recreate computers from first principles.

While similar  in many respects to  the x86 or  ARM ISAs, the LC-3  is radically
simpler  in almost  every  way.  Unlike the  sprawling  x86,  with thousands  of
instructions,  dozens of  addressing  modes, multicore  execution, an  intricate
memory model, and over 40 years of history etched into silicon, the LC-3 remains
a tractable system that is comprehensible by an individual.

It is a  lot closer to a PDP/11  machine than anything you have in  your home or
pocket. Nevertheless, it is still takes quite a lot of effort to understand well
enough to write programs.

-----------------------------
       Project Goals
-----------------------------

𝔼𝕃𝕊𝕀𝔼 is not novel: hardware simulators already exist for the LC-3 architecture,
of course. The textbook publishers provide  one and there are many others freely
available online. This  one is admittedly a mere reinvention  of the wheel. That
said, the gift the design and engineering  process affords is that it can reveal
something  fundamental about  either  our world  or ourselves.  So  it is  worth
retreading the path.

Personally,  there remain  many  Computer  Things that  baffle  me. Despite  ten
thousand hours  of computing, I  still feel  lost when it  comes to some  of the
rudiments of the field:

  - computer architecture;
  - operating systems;
  - assembly programming; and,
  - computing history.

I had a thought that a good way to  learn about these topics was to get my boots
dirty  and learn  the  basics by  building  simple things.  This  project is  an
artifact of  my process. It  is to be  hoped that by trying  my hand at  the old
methods,  by holding  the  craftsman's  tools, by  building  something cute  and
useless, I will gain a better understanding of the essence of computing. Yet, if
nothing else is achieved than learning  a bit, exploring some ideas, and hearing
a few good stories, it will have been worth it.

-----------------------------
       Get in Touch
-----------------------------

I have  lots to learn, many  ideas for experiments to  run, and even a  few more
project plans to bring to my workbench.

You are welcome to reach out if:

  - you're a fellow learner;
  - if you find this project useful or buggy; or,
  - if you have any ideas or questions or feedback.

You can start a discussion on the GitHub project
<https://github.com/smoynes/elsie/discussions> or  you're welcome to  contact me
directly through my GitHub profile.

Please follow  the project  if you  enjoy the absurdist  theatrics of  a curious
software engineer. As ever, I simply seek to understand the essence of computing
and to embody _Shokunin Kishitsu_ (職人気質), the artisan's spirit.

-----------------------------
         Dedication
-----------------------------

This work is dedicated to the MCM/70[2] and its pioneering designers.
<https://en.wikipedia.org/wiki/MCM/70>

-----------------------------
        Documentation
-----------------------------

- TUTORIAL.md          A trailhead for users -- start here.
- DEVGUIDE.txt         Development guidebook.
- RESOURCE.txt         Inspirational references.
- LICENCE.txt          Terms of use.
- CODE_OF_CONDUCT.txt  Behave yourself.

-----------------------------

𝔼𝕃𝕊𝕀𝔼 © 2023 by Scott Moynes is licensed under CC BY-SA 4.0.
See LICENCE.txt for terms. Send your lawyers here:
https://creativecommons.org/licenses/by-sa/4.0/?ref=chooser-v1
