main = { node } .
node = [ doccomment ] { attribute } (decl | stat) .

attribute = "[" attrbody {"," attrbody} "]" .
attrbody  = ident ["=" string] .

stat = deferstat
     | ifstat
     | matchstat
     | loopstat
     | returnstat
     | blockstat
     | callstat
     | assignstat .

deferstat  = "defer" callexpr ";"
ifstat     = "if" expr block {"else" "if" expr block} ["else" block] .
matchstat  = "match" expr "{" { ("_" | expr) "->" stat }  "}" .
loopstat   = "for" [expr] block .
returnstat = "return" [expr] ";" .
blockstat  = block .
callstat   = callexpr ";" .
assignstat = accessexpr "=" expr ";" .

block = "{" {node} "}" .

decl = structdecl
     | usedecl
     | traitdecl
     | impldecl
     | moduledecl
     | funcdecl
     | enumdecl
     | vardecl .

structdecl  = "struct" ident "{" {vardeclbody [","]}  "}" .
usedecl     = "use" ident ";" .
traitdecl   = "trait" ident "{" {funcdecl} "}" .
impldecl    = "impl" ident ["for" ident] "{" funcdecl "}" .
moduledecl  = "module" ident "{" decl "}" .
funcdecl    = "func" ident "(" [funcarg {"," funcarg}] ")" [":" type] ((block | ("->" (stat | expr))) | ";") .
funcarg     = vardeclbody | "..." .
enumdecl    = "enum" ident "{" [enumentry {"," enumentry}] "}" .
enumentry   = ident ["=" expr] .
vardecl     = vardeclbody ";" .
vardeclbody = ["mut"] ident ":" [type] ["=" expr] .

type = pointertype
     | tupletype
     | arraytype
     | typeref .

pointertype = "^" type .
tupletype   = "(" type { "," type } ")" .
arraytype   = "[" [number] "]" type .
typeref     = ident .

expr = primexpr [binop primexpr] .
primexpr = sizeofexpr
         | addrofexpr
         | litexpr
         | castexpr
         | unaryexpr
         | callexpr
         | accessexpr .

sizeofexpr = "sizeof" "(" expr ")" .
addrofexpr = "&" accessexpr .
litexpr    = arraylit
           | tuplelit
           | boollit
           | numlit
           | stringlit
           | runelit .
castexpr   = "cast" "(" type "," expr ")" .
unaryexpr  = unaryop expr .
callexpr   = accessexpr "(" [expr {"," expr}] ")" .
accessexpr = (("^" expr) | ident) {structaccess | arrayaccess | tupleaccess} .

arraylit  = "[" expr {"," expr} "]" .
tuplelit  = "(" expr {"," expr} ")" .
boollit   = "true" | "false" .
numlit    = number .
stringlit = string .
runelit   = rune .

structaccess = "." ident .
arrayaccess  = "[" expr "]" .
tupleaccess  = "|" expr "|" .

binop   = "+" | "-" | "*" | "/" | "%" | ">" | "<" | ">=" | "<=" | "==" | "!="
        | "&" | "|" | "^" | "<<" | ">>" | "&&" | "||" .
unaryop = "!" | "~" | "-" .
