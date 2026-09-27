# file Integerated Programming language or fipl/philip
a structured programming language for the grepless.\
warning: it has a lot of bugs as well it is primary garbage in garbage out so you might not necessarily get an error if you're wrong(even then it might be terribly unheplful)
# design
so have you ever been in a codebase and been like "wait where did they put this code?".\
i have its a dumb question made by a dumb problem over delegation(yes i just made this term up).\
what is over delegation?\
in my eyes its when you put code in places so that you dont have to look at it either because you've overcomplicated things.\
or because invented some boiler for yourself.\
some other people have invented some boiler for you(this is a trap i have fallen into).\
or (rarely) you actually do need to put it in a different file.\
so then you go and you put like 3 functions and 2 structs in 1 file, 4 functions and a struct in another,\
and then oh no! this function relies on this other struct better glue these together recursively.\
philip does not solve this, at least not technically. you still have to deal with a whole bunch of files in you're codebase infact it's worse.\
however you dont have to grep for functions anymore.\
because the functions can be found inside the file browser(yes i have gotten to the point).\
philips main thing is that files are functions and functions are files.\
so you can see the structure of a codebase pretty well without having to look inside a single file(mutual recursion might look a little weird and you do still have to check the includes to see how things connect)
# syntax 
ill explain in depth later but\
ifs like ada\
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
variables are declared starting with the 
```
the a = 1
```
arrays like c and golang but dont have the ability to go across a line\
```
the array = {1,2,3}
```
function calls are normal\
```
fx(1)
```
functions must have their stuff at the top of the file(white space is allowed )\
filename:fx\
```

(input int)
print(input)
```
