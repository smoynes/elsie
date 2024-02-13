;;; Example program using trap GETC.
;;;
;;; Prompt for a character using trap 0x20.
    .ORIG 0x3000
    LD R1,0x0000
    BR START

NEXT_TASK:
    HALT

START:
    GETC                        ; TRAP x20 ; Result is in R0
    BR NEXT_TASK
