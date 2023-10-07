    ;; Fig. 7.1
    .ORIG   x3100
    LD      R1,SIX
    LD      R2,NUMBER
    AND     R3,R3,#0

AGAIN:  ADD R3,R3,R2
        ADD R1,R1,#-1
        BRp AGAIN

    HALT

NUMBER  .BLKW   1
SIX     .FILL   x0006

    .END
