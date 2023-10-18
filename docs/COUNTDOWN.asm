;;; COUNTDOWN.asm : a program for a tutorial that counts down to blast off
;;; Inputs: R0 is the address of counter.

    .ORIG   x3000               ; Start at the beginning of user-space memory.
    ST      R1,SAVER1           ; Set aside the R1 for the counter value.
    LD      R1,COUNT            ; Load the counter value from the pointer.

    ;;
    ;; Main program loop
    ;;
LOOP:
    BRz     EXIT                ; If counter is zero, exit program.
    ADD     R1,R1,#-1           ; Decrement counter.
    BRnzp   LOOP                ; Loop

    ;;
    ;;  Exit program.
    ;;
EXIT:
    LD      R1,SAVER1
    HALT

    ;;
    ;; Program data
    ;;
COUNT:                          ;
    .FILL   10

SAVER1 .DW x0000
    .END
