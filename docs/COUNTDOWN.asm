;;; COUNTDOWN.asm : a program for a tutorial that counts down to blast off
;;; Inputs: R0 is the address of counter.

    .ORIG   x3000               ; Start at the beginning of user-space memory.
    LD      R1,COUNT            ; Load the counter value from the pointer.

    ;;
    ;; Main program loop
    ;;
LOOP:
    BRz     EXIT                ; If counter is zero, exit program.
    ADD     R1,R1,#-1           ; Decrement counter.
    ADD     R0,R1,#xf           ; Add 0x30 to counter and store in R0.
    ADD     R0,R0,#xf           ;
    ADD     R0,R0,#xf           ;
    ADD     R0,R0,#x3           ;
    TRAP    x21                 ; TRAP:  OUT.
    AND     R1,R1,R1            ; Logical check R1 for zero.
    BR      LOOP                ; Loop, again.

    ;;
    ;;  Exit program
    ;;
EXIT:
    HALT

    ;;
    ;; Static data
    ;;
COUNT:
    .FILL   10                  ; Constant to countdown.

    .END
