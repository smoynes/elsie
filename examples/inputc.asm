;;; Example Keyboard Input Echo Program
;;;
;;; Adapted from 3e. Fig 9.7

    .ORIG 0x3000
    LD R1,0x0000
    BR START

NEXT_TASK:
    HALT

START:
    TRAP x20
    BR NEXT_TASK
