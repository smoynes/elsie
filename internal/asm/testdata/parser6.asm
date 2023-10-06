    ;; Fig. 7.2, 2/e.
    .ORIG   x3000
    AND     R2,R2,#0            ; Initialize counter: R2 <- 0
    LD      R3,PTR              ; Pointer to input:   R3 <- 0x4000
    TRAP    x23                 ; Read char input:    R0 <- CHAR
    LDR     R1,R3,#0            ; Fetch next char:    R1 <- [PTR]

TEST:
    ADD     R4,R1,#-4           ; Test for EOT.
    BRz     OUTPUT              ; If true, exit loop: TEST.

    ;; Test for match.
    NOT     R1,R1               ; R1 <- ^R1
    ADD     R1,R1,R0            ; R1 <- R1+R0
    NOT     R1,R1               ; Test match: R1
    BRnp    GETCHAR             ; If not match then do not increment.
    ADD     R2,R2,#1            ; Increment counter:  R2 <- R2+1

GETCHAR:
    ADD     R3,R3,#1            ; Increment pointer: R3 <- R3+1
    LDR     R1,R3,#0            ; Fetch next char:   R1 <- [PTR]
    BRnzp   TEST                ; Loop:              TEST

    ;; Display output.
OUTPUT:
    LD      R0,ASCII            ; Load ASCII text base:     R0 <- x0030
    ADD     R0,R0,R2            ; Add offset to ASCII base: R0 <- R2+0x0030
    TRAP    x21                 ; Print ASCII code:         IO <- R0
    TRAP    x25                 ; Halt and catch fire.

ASCII   .FILL   x0030
PTR     .FILL   x4000

        .END
