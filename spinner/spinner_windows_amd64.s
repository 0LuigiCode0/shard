#include "textflag.h"
   
TEXT ·spin(SB), NOSPLIT, $0-1
   MOVB cycle+0(FP), AX
   CMPB AX, $0
   JZ ret
loop:
   PAUSE
   SUBB $1, AX
   JNZ loop
ret:
   RET
