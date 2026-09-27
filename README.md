# file Integerated Programming language or fipl/philip
a structured programming language for the grepless.\
warning: it has a lot of bugs as well it is primary garbage in garbage out so you might not necessarily get an error if you're wrong(even then it might be terribly unheplful)
# design
so have you ever been in a codebase and been like "wait where did they put this code?".\
i have. its a dumb question. made by a dumb problem called over delegation(yes i just made this term up).\
well philip doesnt solve this , it actually makes it more extreme, which is better hopefully.\
philips main thing is that files are functions and functions are files, and maybe structs are files we'll see.\
so you can see the structure of a codebase pretty well without having to look inside a single file(mutual recursion might look a little weird and you do still have to check the includes to see how things connect)
# syntax 
ill explain in depth later but\
ifs like ada
```
if cond then
#  dothis
else
# dothat
end if
```
for is like ada and golang(still need to implement the i = 0; i<something; i++ thing)\
```
for cond loop
#  dothis until cond is false
end loop
```
variables are declared starting with "the" they do not have type annotation
```
the a = 1
```
arrays like c and golang but dont have the ability to go across a line\
```
the array = {1,2,3}
```
function calls are normal
```
fx(1)
```
functions must have their stuff at the top of the file(white space is allowed )\
args must have type annotation(weird i know)\
filename:fx
```

(input int)
print(input)
```
returns are just types\
filename:fy
```
(input int)(int)
return input
```
