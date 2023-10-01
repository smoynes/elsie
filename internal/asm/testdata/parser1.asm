    ;; Origin
    .ORIG    0x3000

label:
    AND
    AND R1
    AND R1,R2
    AND R1,R2,R3

    ;; Immediate mode: decimal, hex, octal, binary
    AND R1, #1
    AND R1, #x1
    AND R1, #o1
    AND R2, #0b1010_1111

    AND R2, FOO                 ; Symbolic reference
    and R2, [FOO]               ; Indirect

FOO:    0x1234                  ; Data
BAR:    01234
BAZ:    1234
BAT:    '⍣'
STRING: .stringz   "Hi there!"     ; fill string directive
