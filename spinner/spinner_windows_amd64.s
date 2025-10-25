#include "textflag.h"
   
TEXT ·spin(SB), NOSPLIT, $0-4
   MOVL cycle+0(FP), AX
loop:
   PAUSE
   SUBL $1, AX
   JNZ loop
   RET
