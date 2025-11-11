; Test real I/O using memory-mapped addresses
; Output to $FFF0, input from $FFF1

        .ARCH   65c02
        .ORG    $1000

; I/O addresses for our custom memory handler
PUTCH   =       $FFF0           ; Write here to output a character
GETCH   =       $FFF1           ; Read here to get a character

START:
        ; Print "Hello! "
        LDX     #$00
PRINT:
        LDA     MESSAGE,X
        BEQ     DONE
        STA     PUTCH           ; Output character
        INX
        BNE     PRINT

DONE:
        BRK                     ; Halt

MESSAGE:
        .DB     "Hello from real I/O!", $0D, $0A, $00

        .EX     START
