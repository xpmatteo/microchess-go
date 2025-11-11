; Test both input and output
; Read characters from input buffer and echo them to output buffer

        .ARCH   65c02
        .ORG    $1000

; I/O buffer addresses
INBUF   =       $E000           ; Input buffer
OUTBUF  =       $F000           ; Output buffer

START:
        LDX     #$00            ; Initialize index to 0

LOOP:
        LDA     INBUF,X         ; Read character from input buffer
        BEQ     DONE            ; If zero, we're done
        STA     OUTBUF,X        ; Echo to output buffer
        INX                     ; Next character
        CPX     #$40            ; Max 64 chars
        BNE     LOOP

DONE:
        LDA     #$00            ; Add null terminator to output
        STA     OUTBUF,X
        BRK                     ; Halt

        .EX     START
