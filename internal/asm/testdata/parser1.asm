    ;; Origin
    .ORIG    0x3000

    ;; Code section.

label:
    OP
    OP R1
    OP R1,R2
    OP R1,R2,R3

    ;; Immediate mode: decimal, hex, octal.
    OP R1, #1
    OP R1, #0x1
    OP R1, #01

    LDR R2, =FOO                ; Reference
    LDR R2, [FOO]               ; Indirect

FOO:    0x1234                  ; Data
BAR:    01234
BAZ:    1234
BAT:    '⍣'
STRING: .stringz   "Hi there!"     ; fill directive
