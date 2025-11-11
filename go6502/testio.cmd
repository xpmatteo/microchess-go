; Load test program
load testio.bin
; Pre-populate input buffer with test data
ms $E000 $54 $65 $73 $74 $20 $69 $6E $70
ms $E008 $75 $74 $3A $20 $31 $32 $33 $34
ms $E010 $00
; Set PC and run
reg PC START
si 100
; Show input buffer
m $E000 20
; Show output buffer
m $F000 20
quit
