; Minimal test program - print "Hello, MicroChess!" and halt
; This will help us figure out how I/O works in go6502

        .ARCH   65c02
        .ORG    $1000

START:
        LDX     #$00            ; Initialize index to 0
        LDY     #$00            ; Y will be output buffer index

LOOP:
        LDA     MESSAGE,X       ; Load character from message
        BEQ     DONE            ; If zero, we're done
        JSR     PUTCH           ; Print the character
        INX                     ; Next character
        BNE     LOOP            ; Loop (will always branch since we check for 0 above)

DONE:
        BRK                     ; Halt

; Simple character output routine
; Write to sequential memory locations to capture output
PUTCH:
        STA     $F000,Y         ; Write to buffer starting at F000
        INY                     ; Next position
        RTS

MESSAGE:
        .DB     "Hello, MicroChess!", $0D, $0A, $00

        .EX     START
