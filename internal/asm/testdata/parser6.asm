    ;; Fig. 7.2
    .ORIG   x3000
    AND     R2,R2,#0
    LD      R3,PTR
    TRAP    x23
    LDR     R1,R3,#0

TEST:
    ADD     R4,R1,#-4
    BRz     OUTPUT

    NOT     R1,R1
    ADD     R1,R1,#1
    ADD     R1,R1,R0
    BRnp    GETCHAR
    ADD     R2,R2,#1

GETCHAR:
    ADD     R3,R3,#1
    LDR     R1,R3,#0
    BRnzp   TEST

OUTPUT:
    LD      R0,ASCII
    ADD     R0,R0,R2
    TRAP    x21
    TRAP    x25

ASCII  .FILL   x0030
PTR     .FILL   x4000

        .END
