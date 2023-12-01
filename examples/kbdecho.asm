;;; Example Keyboard Input Echo Program
;;;
;;; Adapted from 3e. Fig 9.7

    .ORIG 0x3000
    LD R1,0x0000
    BR START

NEXT_TASK:
    HALT

START:
    ST  R1,SaveR1   ; Save registers needed
    ST  R2,SaveR2   ; by this routine
    ST  R3,SaveR3

    LD R2,Newline
    L1 LDI R3,DSR
    BRzp L1         ; Loop until monitor is ready
    STI R2,DDR      ; Move cursor to new clean line
    LEA R1,Prompt   ; Starting address of prompt string
Loop:
    LDR R0,R1,#0    ; Write the input prompt
    BRz Input       ; End of prompt string
    L2 LDI R3,DSR
    BRzp L2         ; Loop until monitor is ready
    STI R0,DDR      ; Write next prompt character
    ADD R1,R1,#1    ; Increment prompt pointer
    BRnzp Loop      ; Get next prompt character

Input:
    LDI R3,KBSR
    BRzp Input      ; Poll until a character is typed
    LDI R0,KBDR     ; Load input character into R0
    L3 LDI R3,DSR
    BRzp L3         ; Loop until monitor is ready
    STI R0,DDR      ; Echo input character

L4:
    LDI R3,DSR
    BRzp L4         ; Loop until monitor is ready
    STI R2,DDR      ; Move cursor to new clean line
    LD R1,SaveR1    ; Restore registers
    LD R2,SaveR2    ; to original values
    LD R3,SaveR3

    BRnzp NEXT_TASK ; Do the program's next task

SaveR1: .BLKW 1     ; Memory for registers saved
SaveR2: .BLKW 1
SaveR3: .BLKW 1

DSR:        .FILL xFE04         ; Static data
DDR:        .FILL xFE06
KBSR:       .FILL xFE00
KBDR:       .FILL xFE02
Newline:    .FILL x000A ; ASCII code for newline
Prompt:     .STRINGZ "Input any character> "
