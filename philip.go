package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"github.com/message-256/fipl-philip/lexer"
	"strconv"
	"strings"
	"unicode"
	"slices"
	"maps"
)

//i use as<type>()s becuase it simplifys aritmatic
//most if the reason why i use this is becuase generics are dumb for things other that [single any, array []single]
type variable interface {
	asstring() (string,error)
	asint() (int,error)
	getindex([]variable)(variable,error)
	//variable 1 is the index
	setindex([]variable,variable)(error)
	typeof() string
	realtype() string
	retype(string) error
	assign(variable) error
	
}

//interface so big i had to auto generate it
type arithmaticable interface {
	add(variable,variable)variable
	sub(variable,variable)variable
	mul(variable,variable)variable
	div(variable,variable)variable
	not(variable,variable)variable
	and(variable,variable)variable
	or(variable,variable)variable
	shr(variable,variable)variable
	shl(variable,variable)variable
	mod(variable,variable)variable
	lt(variable,variable)variable
	gt(variable,variable)variable
	eqw(variable,variable)variable


}
type variables map[string]*variable

type function struct {
	path string
	args []variable
	returns []variable
	argnames []string
	
}
type functions map[string]function
type interpreter struct {
	funcs functions
	stuff variables
}
type internalString struct{
	value string
	typename string
}
type internalInt struct {
	value int
	typename string
}
type internalArray struct {
	dimensions int
	stuff []variable
	scalar string
}

func (arr* internalArray) asstring() (string,error){
	return "",errors.New("cant get value of type string from int")
}
func (arr* internalArray) asint()(int,error) {
	return 0,errors.New("cannot convert " + arr.typeof() + "to type int")
}
func (arr* internalArray)getindex(index []variable)(variable,error){
	place ,err := index[0].asint()
	if err != nil {
		return nil,err
	}
	if len(index) != 1 {
		 return arr.stuff[place].getindex(index[1:])
	}
	return arr.stuff[place],nil
}
func (arr* internalArray)setindex(index []variable,value variable)( error){
	place ,err := index[0].asint()
	if err != nil {
		return err
	}
	if len(index) != 1 {
		 return arr.stuff[place].setindex(index[1:],value)
	}
	return arr.stuff[place].assign(value)
}
func (arr* internalArray)typeof() string{
	return "[]" + arr.stuff[0].typeof()
}
func (arr* internalArray)realtype() string{
	return "[]"
}
func (arr* internalArray)retype(typename string) error {
	return errors.New("cannot retype array")
}
func (arr* internalArray)assign(input variable) error {
	return errors.New("cant full assign to array after initialization")

}
func (arr* internalArray)append(value variable){
	if value.typeof() != arr.scalar {
		arr.stuff = append(arr.stuff,value)	
	}
}

func (i* internalInt) asstring() (string,error){
	return "",errors.New("cant get value of type string from int")
}
func (i* internalInt) asint()(int,error) {
	return i.value,nil
}
func (i* internalInt)getindex(index []variable)(variable,error){
	return nil,errors.New("cannot index type int")
}
func (i* internalInt)setindex(index []variable,value variable)( error){
	return errors.New("cannot index type int")
}
func (i* internalInt)typeof() string{
	return i.typename
}
func (i* internalInt)realtype() string{
	return "int"
}
func (i* internalInt)retype(newtype string) error {
	i.typename = newtype
	return nil
}
func (i* internalInt)assign(input variable) error {
	var err error
	i.value,err = input.asint()
	return err

}
func (s* internalString) asstring() (string,error){
	return s.value,nil
}
func (s* internalString) asint()(int,error) {
	return 0,errors.New("cant get int from string")
}
func (s* internalString)getindex(index []variable)(variable,error){
	if len(index) > 1 {
		return nil,errors.New(fmt.Sprintf("cannot index %d dimension of type string only has 1 dimension",len(index)))
	}
	place ,err := index[0].asint()
	if err != nil {
		return nil,errors.New("cannot index with type "+index[0].typeof())
	}
	val := s.value[place]
	return &internalInt{
		value:int(val),
		typename:"byte",
	},nil
}
func (s* internalString)setindex(dimensions []variable, value variable)(error){
	return errors.New("strings are read only")
}
func (s* internalString)typeof() string{
	return s.typename
}
func (s* internalString)realtype() string{
	return "string"
}
func (s* internalString)assign(input variable) error {
	if input == nil {
		return errors.New("internal: nil pointer passed to assign")
	}
	var err error
	s.value,err = input.asstring()
	return err

}
func (s* internalString)retype(typename string) error{
	s.typename = typename
	return nil
}
func (i* internalInt) String() string {
	if i == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	return fmt.Sprintf("%d",i.value)
}
func (s* internalString) String() string {
	if s == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	return fmt.Sprintf("%s",s.value)
}
func (arr* internalArray)String() string {
	if arr == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	var collector string
	for i := range arr.stuff {
		collector+=fmt.Sprintf("%v,",arr.stuff[i])
	}
	return collector
	
}

func newInternalInt(i int) variable {
	return &internalInt{
		value:i,
		typename:"int",
	}
}
func newInternalString(s string) variable {
	return &internalString{
		value:s,
		typename:"string",
	}
}
func newInternalArray(v []variable ) variable{
	var name string
	if len(v) > 1 {
		name = v[0].typeof()
	} 
	return &internalArray{
		dimensions:0,
		scalar:name,
		stuff:v,
	}
}
func copyof(v variable) variable {
	switch(v.realtype()){
		case "string":
			s,_ := v.asstring()
			return newInternalString(s)
		case "int":
			i,_ := v.asint()
			return newInternalInt(i)
		case "[]":
			var a1 []variable
			a2,_ := v.(*internalArray)
			for i := range a2.stuff {
				a1 = append(a1,copyof(a2.stuff[i]))
			}
			return newInternalArray(a1)
	}
	return nil
}
func newVariableOfType(typename string)(variable,error) {
	switch(typename){
		case "string":
			return newInternalString(""),nil
		case "int":
			return newInternalInt(0),nil
	}
	if typename[:2] == "[]"{
		v,err := newVariableOfType(typename[:2])
		returned := []variable{v}
		if err != nil {
			return nil,err
		}
		return newInternalArray(returned),nil
	}
	return nil,errors.New("unknown type " + typename)
} 
func bint(b bool) int{
	if b {
		return 1
	}
	return 0
}
func (context interpreter)expr( input []lexer.Ast) (variable, error) {
	if input[0].Type() != "data" && input[0].Type() != "perenthesis" {
		return nil,errors.New("must not start with " + input[0].Value())
	}
	var opperations = map[string]func(int,int) int {
		"/":func(a,b int)int{return a / b},
	 	"*":func(a,b int)int{return a * b},
		"%":func(a,b int)int{return a % b},
	 	"+":func(a,b int)int{return a + b},
		"-":func(a,b int)int{return a - b},
		"<=":func(a,b int)int{return bint(a <= b) },
	 	">=":func(a,b int)int{return bint(a >= b) },
		"==":func(a,b int)int{return bint(a == b) },
		"<":func(a,b int)int{return bint(a < b) },
	 	">":func(a,b int)int{return bint(a > b) },
	}

	var a int
	var oppername string
	var next int
	next = slices.IndexFunc(input,func(a lexer.Ast)bool {return a.Type() == "arithmatic"})
	v ,err := context.valueof(input[:next])
	oppername = input[next].Value()
	if err != nil {
		return nil,err
	}
	a ,err = v.asint()
	if err != nil {
		return nil,err
	}
	input = input[next+1:]	
	for len(input) != 0 {
		v,err := context.valueof(input[:next])
		if err != nil {
			return nil,err
		}
		b,err := v.asint()
		if err != nil {
			return nil,err
		}
		_,ok := opperations[oppername]
		if !ok {
			return nil,errors.New("uncased operation")
		}
 		a = opperations[oppername](a,b)
		next = slices.IndexFunc(input,func(a lexer.Ast)bool {return a.Type() == "arithmatic"})
		if next == -1 {
			next = len(input)-1
		}
		oppername = input[next].Value()
		input = input[next+1:]


	}
	return newInternalInt(a),nil
}
func (context interpreter)call(this []lexer.Ast)([]variable,error) {
	name := this[0].Value()
	args := lexer.SplitFunc(this[1].Inner(),func(a lexer.Ast)bool{return a.Value() == ","})
	f ,ok := context.funcs[name]
	if !ok {
		return nil,fmt.Errorf("function %v does not exist",this[0].Value())
	}
	if len(args) != len(f.args) {
		return nil,fmt.Errorf("(%v) does not stisfy (%v)",args,f.args)
	}
	for i := range f.args {
		v,err := context.valueof(args[i])
		if err != nil {
			return nil,err
		}
		err = f.args[i].assign(v)
		if err != nil {
			return nil,err
		}
	}
	file,err := os.Open(f.path)
	if err != nil {
		return nil,err
	}
	returned := make([]variable,len(f.returns))
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	keyword,err := f.eval(scanner,interpreter{})
	if err != nil {
		return nil,err
	}
	if keyword != "return" && keyword != "" {
		return nil,errors.New("stray " + keyword)
	}
	for i := range f.returns {
		v,err := newVariableOfType(f.returns[i].typeof())
		if err != nil {
			return nil,err
		}
		returned[i] = f.returns[i]
		f.returns[i] = v
	}
	return returned,nil
}
func (context interpreter)valueof(this []lexer.Ast) (variable, error) {
	if len(this) == 0 {
		return nil,nil
	}
	name := this[0].Value();
	if  name[0] == '"' && name[len(name)-1] == '"'{
		return newInternalString(name[1:len(name)-1]),nil
	}
	
	if this[0].Value() == "{}" {
		inner := this[0].Inner()
		split := lexer.SplitFunc(inner,func(a lexer.Ast) bool {return a.Value() == ","})
		firstelement,err := context.valueof(split[0])
		if err != nil {
			return nil,err
		}
		var newarray []variable
		var collecterr error
		var typename = firstelement.typeof()
		for i := range split {
			v,err := context.valueof(split[i])
			if v.typeof() != typename {
				return nil,errors.New(fmt.Sprintf("cant put variable of type %s on array of type []%s",v.typeof(),typename))
			}
			newarray = append(newarray,v)
			errors.Join(collecterr,err)
		}
		if collecterr != nil {
			return nil,collecterr
		}
		return newInternalArray(newarray),nil

	}
	if slices.ContainsFunc(this,func(a lexer.Ast)bool{return a.Type() == "arithmatic"}) {
		return context.expr(this)
	}
	if len(this) > 1 {
		if this[0].Type() == "data" && this[1].Value() == "()" {
			returned ,err := context.call(this)
			if err != nil {
				return nil,err
			}
			if len(returned) != 1 {
				return nil,fmt.Errorf("expected 1 have %d",len(returned))
			}
			return returned[0],nil
		}
		name := this[0]
		this = this[1:]
		if this[0].Value() == "[]" {
			var dimensions []variable
			for i := range this {
				if this[i].Value() != "[]" {
					return nil,fmt.Errorf("stray %v in indexing expression",this[i].Value())
				}
				v,err := context.valueof(this[i].Inner())
				if err != nil {
					return nil,err
				}	
				dimensions = append(dimensions,v)
			}
			indexed ,err := context.valueof([]lexer.Ast{name})
			if err != nil {
				return nil,err
			}
			return indexed.getindex(dimensions)
		}
	}
	if unicode.IsDigit(rune(this[0].Value()[0])) {
		i,err := strconv.Atoi(this[0].Value())
		return newInternalInt(i),err
	} else {
		returned, ok := context.stuff[this[0].Value()]
		
		if !ok {
			return nil, errors.New("variable name not found \"" + this[0].Value() + "\"")
		}
		return copyof(*returned), nil
	}
	return nil, nil
}
func (context interpreter)tuple(this []lexer.Ast)([]variable,error){
	var returned []variable
	if this[0].Type() == "data" && this[1].Value() == "()" {
		return context.call(this)
	} else {
		stuff := lexer.Split(this,lexer.Looker{Value:","})
		var collective error
		for i := range stuff {
			v,err := context.valueof(stuff[i])
			if err != nil {
				collective = errors.Join(collective,err)
				continue	
			}
			returned = append(returned,v)
		}
		if collective != nil {
			return nil,collective
		}
	}
	return returned,nil
}
func (context variables)New(this lexer.Ast,initial variable) error {
	if unicode.IsDigit(rune(this.Value()[0])){
		return errors.New(fmt.Sprint(this,":cant start variable name with number."))

	}
	context[this.Value()] = new(variable)
	*context[this.Value()] = initial
	return nil
}
func itlookslikeassignment(ast []lexer.Ast) bool {
	return slices.ContainsFunc(ast,func(a lexer.Ast) bool {return a.Value() == "="})
}
func (context function)eval(scanner *bufio.Scanner,usages interpreter)(keyword string,err error){
 	if usages.stuff == nil {
		usages.stuff = make(variables)
	}
	if usages.funcs == nil {
		usages.funcs = make(map[string]function)
	}
	if usages.funcs == nil {
		usages.funcs = make(map[string]function)
	}
	if usages.funcs == nil {
		usages.funcs = make(map[string]function)
	}
	for i:= range context.args {
		usages.stuff[context.argnames[i]] = new(variable)
		*usages.stuff[context.argnames[i]] = context.args[i]
	}
	var ast []lexer.Ast
	for scanner.Scan() {
		line := scanner.Text()
		ast,err = lexer.Lex(line)
		if err != nil {
			return
		}
		if ast == nil {
		} else if len(ast) < 2 && ast[0].Value() != "return" && ast[0].Value() != "break" && ast[0].Value() != "continue"{
			err = errors.New("cant think of any reason to have a like with less that 2 tokens on it")
			return
		} else if ast[0].Value() == "break" {
			keyword = "break"
			return	
		} else if ast[0].Value() == "continue" {
			keyword = "continue"
			return	
		} else if ast[0].Value() == "return" {
			if len(context.returns) > 0 && len(ast) < 2 {
				err = fmt.Errorf("cannot return nil from function with %v",context.returns)
				return
			}
			keyword = "return"		 
			names := lexer.Split(ast[1:],lexer.Looker{Value:","})
			if len(context.returns) != len(names){
				err = errors.New(fmt.Sprintf("%v does not satisfy %v",names,context.returns))
			}
			var v variable
			for i := range context.returns {
				v,err  = usages.valueof(names[i])
				if err != nil {
					return
				}
				err = context.returns[i].assign(v)
				if err != nil {
					return
				}
			}
			return 
		} else if ast[0].Value() == "include" {
			namestring := ast[1].Value()
			name := namestring[1:len(namestring)-1]
			var thenew function
			file,err2 := os.Open(name)
			if err2 != nil {
				err = err2
				return
			}
			thenew.path = name
			argchecker := bufio.NewScanner(file)
			var args []lexer.Ast
			for argchecker.Scan() && args == nil {
				args,err = lexer.Lex(argchecker.Text())
				if err != nil {
					return
				}
			}
			if args[0].Type() == "perenthesis" {
				if len(args) == 2 {
					inner := lexer.SplitFunc(args[1].Inner(),func(a lexer.Ast)bool{return a.Value() == ","}) 			 	
					var v variable
					for i := range inner {
						v,err = newVariableOfType(inner[i][0].Value())
						if err != nil {
							return
						}
						thenew.returns = append(thenew.returns,v)
					}	
				}
				inner := lexer.SplitFunc(args[0].Inner(),func(a lexer.Ast)bool{return a.Value() == ","}) 			 	
				var v variable
				for i := range inner {
					thenew.argnames = append(thenew.argnames,inner[i][0].Value())
					v,err = newVariableOfType(inner[i][1].Value())
					if err != nil {
						return
					}
					thenew.args = append(thenew.args,v)
				}
					
			} else {
				err = errors.New("first declaration must be args")
				return
			}
			usages.funcs[name] = thenew
		} else if ast[0].Value() == "print" && ast[1].Value() == "()"{
			inner := ast[1].Inner()
			split := lexer.SplitFunc(inner,func(a lexer.Ast)bool{return a.Value() == ","})
			var printed []variable
			var collective error
			for i := range split {
				v,err := usages.valueof(split[i])	
				printed = append(printed,v)
				collective = errors.Join(collective,err)
			}
			if collective != nil {
				fmt.Println(collective)
			} else {
				fmt.Printf("%v\n",printed)
			}
		} else if ast[0].Value() == "the" {
			n := slices.IndexFunc(ast,func(ast lexer.Ast)bool{return ast.Value() == "="})
			assignedtolist := lexer.Split(ast[1:n],lexer.Looker{Value:","})
			if len(assignedtolist) == 1 {
				var val variable
				val ,err = usages.valueof(ast[n+1:])
				if err != nil {
					return
				}
				err = usages.stuff.New(ast[1],val)
			} else {
				var vals []variable
				
				vals,err = usages.tuple(ast[n+1:])
				if err != nil {
					return 
				}
				if len(vals) != len(assignedtolist) {
					err = fmt.Errorf("%v could not satisfy %v",assignedtolist,ast[n+1])
					return
				}
				for i := range assignedtolist {
					err = errors.Join(err,usages.stuff.New(assignedtolist[i][0],vals[i]))

				}
			}
			if err != nil {
				fmt.Println(assignedtolist,":",err)
			}
		} else if itlookslikeassignment(ast) {
			n := slices.IndexFunc(ast,func(ast lexer.Ast)bool{return ast.Value() == "="})
			if n >= len(ast)-1 {
				err = fmt.Errorf("%v does nothing",ast)
			}
			varnames := lexer.SplitFunc(ast[:n],func(a lexer.Ast)bool{return a.Value() == ","})
			var v []variable
			if len(varnames) == 1 {
				v = make([]variable,1)
				v[0],err = usages.valueof(ast[n+1:])	
			} else {
				v,err = usages.tuple(ast[n+1:])
			}
			if err != nil {
				return
			}
			if len(varnames) != len(v) {
				err = fmt.Errorf("%v does not satisfy %v",v,varnames)
				return
			}
			var collective error
			for i := range varnames {
				if len(varnames[i]) == 1 {
					place,ok := usages.stuff[varnames[i][0].Value()]
					if ok {
						err = (*place).assign(v[i])
					} else {
						err = errors.New(varnames[i][0].Value() + ":not found ")
					}
				
				} else {
					
					name := varnames[i][0].Value()
					varnames[i] = varnames[i][1:]
					var index []variable
					var val variable 
					for j := range varnames[i] {
						if varnames[i][j].Value() != "[]" {
							err = fmt.Errorf("stray %v",varnames[i][j])
							return
						}
						val,err = usages.valueof(varnames[i][j].Inner())
						if err != nil {
							return
						}
						index = append(index,val)
					}
					if err != nil {
						fmt.Println(err)
						continue
					}
					place,ok := usages.stuff[name]
					if ok {
						err = (*place).setindex(index,v[i])
					} else {
						err = errors.New(name + ":not found ")
					}	
				}
			}
		 	if collective != nil {
				err = collective
				return
			}
		} else if ast[0].Value() == "if" && ast[len(ast)-1].Value() == "then"{
			if len(ast) == 2 {
				err = errors.New("if but no condition")
				return
			}
			var condition int
			var v variable
			v,err = usages.expr(ast[1:len(ast)-1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			condition ,err = v.asint()
			if err != nil {
				fmt.Println(err)
				continue
			}
			var ifLambda,elseLambda string
			var lambda *string = &ifLambda
			var collective error
			var sp int = 1
			for scanner.Scan() {
				line = scanner.Text()
				ast ,err := lexer.Lex(line)
				if ast != nil {
					if ast[0].Value() == "end" && ast[len(ast)-1].Value() == "if"{
						sp--
					} else if ast[0].Value() == "if" && ast[len(ast)-1].Value() == "then" {
						sp++
					} else if ast[0].Value() == "else" && sp == 1 {
						lambda = &elseLambda
						continue
					}
				}
				collective = errors.Join(collective,err)
				if sp == 0 {
					break
				}
				*lambda+=line+"\n"
				

			}
			if collective != nil {
				err = collective
				return
			}
			if condition >= 1 {
 				keyword,err = context.eval(bufio.NewScanner(strings.NewReader(ifLambda)),interpreter{stuff:maps.Clone(usages.stuff),funcs:maps.Clone(usages.funcs)})
				if keyword == "return" {
					return
				}
				if err != nil {
					fmt.Println(err)
				}
			} else {
				keyword,err = context.eval(bufio.NewScanner(strings.NewReader(elseLambda)),interpreter{stuff:maps.Clone(usages.stuff),funcs:maps.Clone(usages.funcs)})
				if keyword == "return" {
					return
				}
				if err != nil {
					fmt.Println(err)
				}
			}
		} else if ast[0].Value() == "for" && ast[len(ast)-1].Value() == "loop" {
			expression := ast[1:len(ast)-1]		
			var lambda string
			var sp int = 1
			var collective error
			for scanner.Scan() {
				line = scanner.Text()
				ast ,err := lexer.Lex(line)
				if ast != nil && len(ast) >= 2 {
					if ast[0].Value() == "end" && ast[len(ast)-1].Value() == "loop"{
						sp--
					} else if ast[0].Value() == "for" && ast[len(ast)-1].Value() == "loop" {
						sp++
					}
				}
				collective = errors.Join(collective,err)
				if sp == 0 {
					break
				}
				lambda+=line+"\n"
				

			}
			if !(ast[0].Value() == "end" && ast[len(ast)].Value() == "loop") {
					fmt.Println("loop but no end")
			}
			var returnedkeyword string
			var condition int
			var f function
			for {
				v,err := usages.expr(expression)
				if err != nil {
					fmt.Println(err)
					break
				}
				
				condition,err = v.asint()
				if err != nil {
					fmt.Println(err)
					break
				}
	 			returnedkeyword,err = f.eval(bufio.NewScanner(strings.NewReader(lambda)),interpreter{stuff:maps.Clone(usages.stuff),funcs:maps.Clone(usages.funcs)})
				switch(returnedkeyword){
					case "break":
					break
					case "continue":
					
				}
				if err != nil {
					fmt.Println(err)
					break;
				}
				if condition <= 0 {
					break;
				}
			}
		} else if ast[0].Type() == "data" && ast[1].Value() == "()" {
			var returned []variable
			returned,err = usages.call(ast)
			if err != nil {
				return "",err
			}
			if len(returned) != 0 {
				return "",errors.New("non procedure call to procedure")
			}
		} else {
			fmt.Println("error",line,"looks like gibberish")
		}
	}
	return
}
func main() {
	var file *os.File
	var err error
	if len(os.Args) > 1 {
		file, err = os.Open(os.Args[1])
	} else {
		file = os.Stdin
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var f function
	var keyword string
	keyword,err = f.eval(bufio.NewScanner(file),interpreter{})
	if err != nil {
		fmt.Println(err)
	}
	if keyword != "return" && keyword != "" {
		fmt.Println("error stray",keyword)
	}

}
