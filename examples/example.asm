      .ORIG x3000
      LD   R1,COUNT
      LD   R2,ASCII
LOOP  BRz  EXIT
      ADD  R0,R2,R1
      TRAP x21
      ADD  R1,R1,#-1
      BR   LOOP
EXIT  HALT
COUNT .DW  5
ASCII .DW  48
      .END
