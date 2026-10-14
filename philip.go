package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/message-256/fipl-philip/lexer"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type variable interface {
	getindex([]variable) (variable, error)
	setindex([]variable, variable) error
	typeof() string
	realtype() string
	retype(string) error
	assign(variable) error
	len() (variable, error)
	append(variable) (variable, error)
}

type variables map[string]*variable

type function struct {
	dir      string
	path     string
	args     []variable
	returns  []variable
	argnames []string
}
type structure struct {
	vars map[string]string
}
type functions map[string]function
type interpreter struct {
	packagename string
	packaged    map[string]functions
	structs     map[string]structure
	funcs       functions
	stuff       variables
}

// field is essentially map
// hope this doesnt cuase issues later
type internalStruct struct {
	fields   map[string]variable
	typename string
}
type internalString struct {
	value    string
	typename string
}
type internalInt struct {
	value    int
	typename string
}
type internalArray struct {
	dimensions int
	stuff      []variable
	scalar     string
}
type internalFloat struct {
	value    float32
	typename string
}

func (st *internalStruct) len() (variable, error) {
	return nil, fmt.Errorf("cant len %v its a struct", st.typename)
}
func (arr *internalArray) len() (variable, error) {
	return newInternalInt(len(arr.stuff)), nil
}
func (s *internalString) len() (variable, error) {
	return newInternalInt(len(s.value)), nil
}
func (f *internalFloat) len() (variable, error) {
	return nil, fmt.Errorf("cant len float")
}
func (i *internalInt) len() (variable, error) {
	return nil, errors.New("cant len scalar")
}

func (st *internalStruct) append(v variable) (variable, error) {
	return nil, fmt.Errorf("cant append %v its a struct", st.typename)

}
func (arr *internalArray) append(v variable) (variable, error) {
	if v.typeof() != arr.scalar {
		return nil, fmt.Errorf("cannot append %v to array of type %v", v.typeof(), arr.scalar)
	}
	return newInternalArray(append(arr.stuff, v)), nil
}
func (s *internalString) append(v variable) (variable, error) {
	s2, err := asstring(v)
	if err != nil {
		return nil, err
	}
	return newInternalString(s.value + s2), nil
}
func (f *internalFloat) append(v variable) (variable, error) {
	return nil, fmt.Errorf("cant append to float")
}
func (i *internalInt) append(v variable) (variable, error) {
	return nil, errors.New("cannot append to scalar")
}

func (st *internalStruct) getindex(index []variable) (variable, error) {

	field, err := asstring(index[0])
	if err != nil {
		return nil, fmt.Errorf("index[0].asstring returned error %v in call to getfield() = getindex() might be a bug")
	}
	_, ok := st.fields[field]
	if !ok {
		return nil, fmt.Errorf("struct %v does not have field %v", st.typename, field)
	}
	if len(index) > 1 {
		return st.fields[field].getindex(index[1:])
	}
	return copyof(st.fields[field]), nil
}
func (arr *internalArray) getindex(index []variable) (variable, error) {
	place, err := asint(index[0])
	if err != nil {
		return nil, err
	}
	if len(index) != 1 {
		return arr.stuff[place].getindex(index[1:])
	}
	return copyof(arr.stuff[place]), nil
}
func (s *internalString) getindex(index []variable) (variable, error) {
	if len(index) > 1 {
		return nil, errors.New(fmt.Sprintf("cannot index %d dimension of type string only has 1 dimension", len(index)))
	}
	place, err := asint(index[0])
	if err != nil {
		return nil, errors.New("cannot index with type " + index[0].typeof())
	}
	val := s.value[place]
	return &internalInt{
		value:    int(val),
		typename: "byte",
	}, nil
}
func (f *internalFloat) getindex(index []variable) (variable, error) {
	return nil, fmt.Errorf("cant index type float")
}
func (i *internalInt) getindex(index []variable) (variable, error) {
	return nil, errors.New("cannot index type int")
}

func (st *internalStruct) setindex(index []variable, value variable) error {
	if len(index) > 1 {
		return fmt.Errorf("%v len > 1 might be a bug", index)
	}
	field, err := asstring(index[0])
	if err != nil {
		return fmt.Errorf("index[0].asstring returned error %v in call to getfield() = getindex() might be a bug")
	}
	_, ok := st.fields[field]
	if !ok {
		return fmt.Errorf("struct %v does not have field %v", st.typename, field)
	}

	return st.fields[field].assign(value)
}
func (arr *internalArray) setindex(index []variable, value variable) error {
	place, err := asint(index[0])
	if err != nil {
		return err
	}
	if len(index) != 1 {
		return arr.stuff[place].setindex(index[1:], value)
	}
	return arr.stuff[place].assign(value)
}
func (s *internalString) setindex(dimensions []variable, value variable) error {
	return errors.New("strings are read only")
}
func (f *internalFloat) setindex(index []variable, value variable) error {
	return fmt.Errorf("cant index type float")
}
func (i *internalInt) setindex(index []variable, value variable) error {
	return errors.New("cannot index type int")
}

func (st *internalStruct) typeof() string {
	return st.typename
}
func (arr *internalArray) typeof() string {
	return "[]" + arr.stuff[0].typeof()
}
func (s *internalString) typeof() string {
	return s.typename
}
func (f *internalFloat) typeof() string {
	return f.typename
}
func (i *internalInt) typeof() string {
	return i.typename
}

func (st *internalStruct) realtype() string {
	var name string = "struct{"
	for i := range st.fields {
		name += fmt.Sprintf("%v,", st.fields[i].typeof())
	}
	return name + "}"
}
func (arr *internalArray) realtype() string {
	return "[]"
}
func (s *internalString) realtype() string {
	return "string"
}
func (f *internalFloat) realtype() string {
	return "float"
}
func (i *internalInt) realtype() string {
	return "int"
}

func (st *internalStruct) retype(typename string) error {
	return errors.New("cannot retype struct")
}
func (arr *internalArray) retype(typename string) error {
	return errors.New("cannot retype array")
}
func (s *internalString) retype(typename string) error {
	s.typename = typename
	return nil
}
func (f *internalFloat) retype(newtype string) error {
	f.typename = newtype
	return nil
}
func (i *internalInt) retype(newtype string) error {
	i.typename = newtype
	return nil
}

func (st *internalStruct) assign(input variable) error {
	if input.typeof() == st.typeof() {
		for i := range st.fields {
			othersfield, err := input.getindex([]variable{newInternalString(i)})
			if err != nil {
				return err
			}
			err = st.fields[i].assign(othersfield)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("cant assign type %v to type %v", input.typeof(), st.typeof())
}
func (arr *internalArray) assign(input variable) error {
	if arr.typeof() == input.typeof() {
		newarr, _ := input.(*internalArray)
		arr.stuff = newarr.stuff
		return nil
	}
	return fmt.Errorf("cant assign type %v to type %v", arr.typeof(), input.typeof())

}
func (s *internalString) assign(input variable) error {
	if input == nil {
		return errors.New("internal: nil pointer passed to assign")
	}
	var err error
	s.value, err = asstring(input)
	return err

}
func (f *internalFloat) assign(input variable) error {
	f2, ok := input.(*internalFloat)
	if !ok {
		return fmt.Errorf("cant assign typr %v to float", input.typeof())
	}
	f.value = f2.value
	return nil
}
func (i *internalInt) assign(input variable) error {
	var err error
	i.value, err = asint(input)
	return err
}

func (arr *internalArray) String() string {
	if arr == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	var collector string
	for i := range arr.stuff {
		collector += fmt.Sprintf("%v,", arr.stuff[i])
	}
	return collector

}
func (f *internalFloat) String() string {
	if f == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	return fmt.Sprintf("%v", f.value)
}
func (i *internalInt) String() string {
	if i == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	return fmt.Sprintf("%d", i.value)
}
func (s *internalString) String() string {
	if s == nil {
		return fmt.Sprintf("error:internal nil variable ")
	}
	return fmt.Sprintf("%s", s.value)
}

func newInternalInt(i int) variable {
	return &internalInt{
		value:    i,
		typename: "int",
	}
}
func newInternalFloat(f float32) variable {
	return &internalFloat{
		value:    f,
		typename: "float",
	}
}
func newInternalString(s string) variable {
	return &internalString{
		value:    s,
		typename: "string",
	}
}
func newInternalArray(v []variable) variable {
	var name string
	if len(v) > 0 {
		name = v[0].typeof()
	}
	return &internalArray{
		dimensions: 0,
		scalar:     name,
		stuff:      v,
	}
}
func copyof(v variable) variable {
	switch v.realtype() {
	case "string":
		s, _ := asstring(v)
		return newInternalString(s)
	case "int":
		i, _ := asint(v)
		return newInternalInt(i)
	case "[]":
		var a1 []variable
		a2, _ := v.(*internalArray)
		for i := range a2.stuff {
			a1 = append(a1, copyof(a2.stuff[i]))
		}
		return newInternalArray(a1)
	}
	return nil
}
func newVariableOfTypeName(typename string) (variable, error) {
	switch typename {
	case "string":
		return newInternalString(""), nil
	case "int":
		return newInternalInt(0), nil
	}
	return nil, errors.New("unknown type " + typename)
}
func newVariableOfType(a []lexer.Ast) (variable, error) {
	if len(a) == 1 {
		return newVariableOfTypeName(a[0].Value())
	}
	if len(a) >= 2 {
		if a[0].Value() == "[]" {

			v, err := newVariableOfType(a[1:])
			returned := []variable{v}
			if err != nil {
				return nil, err
			}
			return newInternalArray(returned), nil
		}
	}
	return nil, errors.New("nil type")
}
func bint(b bool) int {
	if b {
		return 1
	}
	return 0
}
func bfloat(b bool) float32 {
	if b {
		return 1
	}
	return 0
}
func asint(v variable) (int, error) {
	i, ok := v.(*internalInt)
	if !ok {
		return 0, fmt.Errorf("cant convert %v to type int", v.typeof())
	}
	return i.value, nil
}
func asstring(v variable) (string, error) {
	s, ok := v.(*internalString)
	if !ok {
		return "", fmt.Errorf("cant convert %v to type string", v.typeof())
	}
	return s.value, nil
}
func asfloat(v variable) (float32, error) {
	f, ok := v.(*internalFloat)
	if !ok {
		return 0, fmt.Errorf("cant convert %v to type float", v.typeof())
	}
	return f.value, nil
}
func (context interpreter) fexpr(this []lexer.Ast) (variable, error) {
	if this[0].Type() != "data" && this[0].Type() != "perenthesis" {
		return nil, errors.New("must not start with " + this[0].Value())
	}
	var a float32
	var oppername string
	var next int
	next = slices.IndexFunc(this, func(a lexer.Ast) bool { return a.Type() == "arithmatic" && a.Value() != "." })
	v, err := context.valueof(this[:next])
	oppername = this[next].Value()
	if err != nil {
		return nil, err
	}
	expressionT := v.typeof()
	a, err = asfloat(v)
	if err != nil {
		return nil, err
	}
	this = this[next+1:]
	for len(this) != 0 {
		v, err := context.valueof(this[:next])
		if err != nil {
			return nil, err
		}
		if v.typeof() != expressionT {
			return nil, fmt.Errorf("cant convert type %v to %v", v.typeof(), expressionT)
		}
		b, err := asfloat(v)
		if err != nil {
			return nil, err
		}
		switch oppername {
		case "/":
			a = a / b
		case "*":
			a = a * b
		case "+":
			a = a + b
		case "-":
			a = a - b
		case "<=":
			a = bfloat(a <= b)
		case ">=":
			a = bfloat(a >= b)
		case "==":
			a = bfloat(a == b)
		case "!=":
			a = bfloat(a != b)
		case "<":
			a = bfloat(a < b)
		case ">":
			a = bfloat(a > b)
		default:
			return nil, fmt.Errorf("uncased operation %v", oppername)
		}
		next = slices.IndexFunc(this, func(a lexer.Ast) bool { return a.Type() == "arithmatic" && a.Value() != "." })
		if next == -1 {
			next = len(this) - 1
		}
		oppername = this[next].Value()
		this = this[next+1:]

	}
	return newInternalFloat(a), nil
}

func (context interpreter) iexpr(this []lexer.Ast) (variable, error) {
	if this[0].Type() != "data" && this[0].Type() != "perenthesis" {
		return nil, errors.New("must not start with " + this[0].Value())
	}
	var a int
	var oppername string
	var next int
	next = slices.IndexFunc(this, func(a lexer.Ast) bool { return a.Type() == "arithmatic" })
	v, err := context.valueof(this[:next])
	oppername = this[next].Value()
	if err != nil {
		return nil, err
	}
	expressionT := v.typeof()
	a, err = asint(v)
	if err != nil {
		return nil, err
	}
	this = this[next+1:]
	for len(this) != 0 {
		v, err := context.valueof(this[:next])
		if err != nil {
			return nil, err
		}
		if v.typeof() != expressionT {
			return nil, fmt.Errorf("cant convert type %v to %v", v.typeof(), expressionT)
		}
		b, err := asint(v)
		if err != nil {
			return nil, err
		}
		switch oppername {
		case "/":
			a = a / b
		case "*":
			a = a * b
		case "%":
			a = a % b
		case "+":
			a = a + b
		case "-":
			a = a - b
		case "<=":
			a = bint(a <= b)
		case ">=":
			a = bint(a >= b)
		case "==":
			a = bint(a == b)
		case "!=":
			a = bint(a != b)
		case "<":
			a = bint(a < b)
		case ">":
			a = bint(a > b)
		default:
			return nil, fmt.Errorf("uncased operation %v", oppername)
		}
		next = slices.IndexFunc(this, func(a lexer.Ast) bool { return a.Type() == "arithmatic" })
		if next == -1 {
			next = len(this) - 1
		}
		oppername = this[next].Value()
		this = this[next+1:]

	}
	return newInternalInt(a), nil
}

// funcs like this arent my thing
// but i didnt feel like changing how expr does stuff
func (context interpreter) expr(this []lexer.Ast) (variable, error) {
	i := slices.IndexFunc(this, func(a lexer.Ast) bool { return a.Type() == "arithmatic" && a.Value() != "." })
	if i == -1 {
		return nil, fmt.Errorf("something weird found its way to expr")
	}
	v, err := context.valueof(this[:i])
	if err != nil {
		return nil, err
	}
	if v.realtype() == "float" {
		return context.fexpr(this)
	} else if v.realtype() == "int" {
		return context.iexpr(this)
	}
	return nil, fmt.Errorf("cant do arithmatic with type %v", v.typeof())
}
func (context interpreter) call(this []lexer.Ast) ([]variable, error) {
	var perens []lexer.Ast
	var ok bool
	var f function
	if this[0].Type() == "data" && this[1].Value() == "()" {
		name := this[0].Value()
		f, ok = context.funcs[name]
		if !ok {
			return nil, fmt.Errorf("function %v does not exist", this[0].Value())
		}
		perens = this[1].Inner()
	} else if this[0].Type() == "data" && this[1].Value() == "." && this[2].Type() == "data" && this[3].Value() == "()" {
		perens = this[3].Inner()
		name := this[2].Value()
		selection := this[0].Value()
		p, ok := context.packaged[selection]
		if !ok {
			return nil, fmt.Errorf("package '%v' does not exist", selection)
		}
		f, ok = p[name]
		if !ok {
			return nil, fmt.Errorf("package %v does not provide %v", selection, name)
		}

	}
	args, err := context.tuple(perens)
	if err != nil {
		return nil, err
	}

	if len(args) != len(f.args) {
		return nil, fmt.Errorf("(%v) does not stisfy (%v)", args, f.args)
	}
	for i := range f.args {
		err = f.args[i].assign(args[i])
		if err != nil {
			return nil, err
		}
	}
	normal, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if normal != f.dir {
		err = os.Chdir(f.dir)
		if err != nil {
			return nil, err
		}
	}
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	returned := make([]variable, len(f.returns))
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ast, err := lexer.Lex(scanner.Text())
		if err != nil {
			return nil, err
		}
		if ast != nil {
			break
		}
	}

	keyword, line, err := f.eval(scanner, interpreter{})
	if err != nil {
		return nil, fmt.Errorf("%v:%v:error %v", f.path, line, err)
	}
	err = os.Chdir(normal)
	if err != nil {
		return nil, err
	}
	if keyword != "return" && keyword != "" {
		return nil, errors.New("stray " + keyword)
	}
	for i := range f.returns {
		v, err := newVariableOfTypeName(f.returns[i].typeof())
		if err != nil {
			return nil, err
		}
		returned[i] = f.returns[i]
		f.returns[i] = v
	}
	return returned, nil
}

func (context interpreter) valueof(this []lexer.Ast) (variable, error) {
	if len(this) == 0 {
		return nil, nil
	}
	if len(this) == 1 && this[0].Value() == "()" {
		return context.valueof(this[0].Inner())
	}
	name := this[0].Value()
	if itlookslikeastring(this[0]) {
		return newInternalString(name[1 : len(name)-1]), nil
	}
	if this[0].Value() == "{}" {
		inner := this[0].Inner()
		split := lexer.SplitFunc(inner, func(a lexer.Ast) bool { return a.Value() == "," })
		firstelement, err := context.valueof(split[0])
		if err != nil {
			return nil, err
		}
		var newarray []variable
		var collecterr error
		var typename = firstelement.typeof()
		for i := range split {
			v, err := context.valueof(split[i])
			if v.typeof() != typename {
				return nil, errors.New(fmt.Sprintf("cant put variable of type %s on array of type []%s", v.typeof(), typename))
			}
			newarray = append(newarray, v)
			errors.Join(collecterr, err)
		}
		if collecterr != nil {
			return nil, collecterr
		}
		return newInternalArray(newarray), nil

	}
	if slices.ContainsFunc(this, func(a lexer.Ast) bool { return a.Type() == "arithmatic" && a.Value() != "." }) || this[0].Value() == "()" {
		return context.expr(this)
	}
	if itlookslikeanattribute(this) {
		if unicode.IsDigit(rune(this[0].Value()[0])) && unicode.IsDigit(rune(this[2].Value()[0])) {
			f, err := strconv.ParseFloat(this[0].Value()+this[1].Value()+this[2].Value(), 32)
			if err != nil {
				return nil, err
			}
			return newInternalFloat(float32(f)), nil
		}
	}

	if itlookslikeafunctioncall(this) {
		returned, err := context.call(this)
		if err != nil {
			return nil, err
		}
		if len(returned) != 1 {
			return nil, fmt.Errorf("expected 1 have %d", len(returned))
		}
		return returned[0], nil
	}

	if len(this) > 1 {
		if this[0].Value() == "len" && this[1].Value() == "()" {
			v, err := context.valueof(this[1].Inner())
			if err != nil {
				return nil, err
			}
			if v == nil {
				return nil, errors.New("len nil value")
			}
			return v.len()
		}
		if this[0].Value() == "append" && this[1].Value() == "()" {
			v, err := context.tuple(this[1].Inner())
			if err != nil {
				return nil, err
			}
			if v == nil {
				return nil, errors.New("append nil value")
			}
			return v[0].append(v[1])
		}

		if this[0].Value() == "[]" {
			var dimensions []variable
			for i := range this {
				if this[i].Value() != "[]" {
					return nil, fmt.Errorf("stray %v in indexing expression", this[i].Value())
				}
				v, err := context.valueof(this[i].Inner())
				if err != nil {
					return nil, err
				}
				dimensions = append(dimensions, v)
			}
			indexed, err := context.valueof([]lexer.Ast{this[0]})
			if err != nil {
				return nil, err
			}
			return indexed.getindex(dimensions)
		}

	}
	if unicode.IsDigit(rune(this[0].Value()[0])) {
		i, err := strconv.Atoi(this[0].Value())
		return newInternalInt(i), err
	} else {
		returned, ok := context.stuff[this[0].Value()]
		if !ok {
			return nil, errors.New("variable name not found \"" + this[0].Value() + "\"")
		}
		return copyof(*returned), nil
	}
	return nil, nil
}
func (context interpreter) tuple(this []lexer.Ast) ([]variable, error) {
	var returned []variable

	if itlookslikeafunctioncall(this) {
		return context.call(this)
	}
	stuff := lexer.Split(this, lexer.Looker{Value: ","})
	var collective error
	for i := range stuff {
		v, err := context.valueof(stuff[i])
		if err != nil {
			collective = errors.Join(collective, err)
			continue
		}
		returned = append(returned, v)
	}
	if collective != nil {
		return nil, collective
	}

	return returned, nil
}
func (context interpreter) refer() interpreter {
	return interpreter{
		stuff: maps.Clone(context.stuff),
		funcs: maps.Clone(context.funcs), 
		packaged: maps.Clone(context.packaged),
	}
}
func (context variables) New(this lexer.Ast, initial variable) error {
	if unicode.IsDigit(rune(this.Value()[0])) || this.Type() != "data" {
		return fmt.Errorf("%v:cant start variable name with number.",this)

	}
	context[this.Value()] = new(variable)
	*context[this.Value()] = initial
	return nil
}
func (context interpreter) newfunc(dir, name string) (function, error) {
	normal, err := os.Getwd()
	if err != nil {
		return function{}, err
	}
	err = os.Chdir(dir)
	if err != nil {
		return function{}, err
	}
	var thenew function
	file, err := os.Open(name)

	thenew.dir, err = os.Getwd()
	if err != nil {
		return function{}, err
	}
	err = os.Chdir(normal)
	if err != nil {
		return function{}, err
	}
	if err != nil {
		return function{}, err
	}
	thenew.path = name
	argchecker := bufio.NewScanner(file)
	var args []lexer.Ast
	for argchecker.Scan() && args == nil {
		args, err = lexer.Lex(argchecker.Text())
		if err != nil {
			return function{}, err
		}
	}
	if args[0].Value() == "()" {
		if len(args) == 2 {
			inner := lexer.SplitFunc(args[1].Inner(), func(a lexer.Ast) bool { return a.Value() == "," })
			for i := range inner {
				v, err := newVariableOfType(inner[i])
				if err != nil {
					return function{}, err
				}
				thenew.returns = append(thenew.returns, v)
			}
		}
		inner := lexer.SplitFunc(args[0].Inner(), func(a lexer.Ast) bool { return a.Value() == "," })
		for i := range inner {
			thenew.argnames = append(thenew.argnames, inner[i][0].Value())
			v, err := newVariableOfType([]lexer.Ast{inner[i][1]})
			if err != nil {
				return function{}, err
			}
			thenew.args = append(thenew.args, v)
		}

	} else {
		return function{}, fmt.Errorf("%v first declaration must be args", name)
	}
	return thenew, nil
}
func itlookslikeassignment(ast []lexer.Ast) bool {
	return slices.ContainsFunc(ast, func(a lexer.Ast) bool { return a.Value() == "=" })
}
func itlookslikeafunctioncall(ast []lexer.Ast) bool {
	if len(ast) >= 2 {
		return ast[len(ast)-2].Type() == "data" && ast[len(ast)-1].Value() == "()"
	}
	return false
}
func itlookslikeanattribute(ast []lexer.Ast) bool {
	if len(ast) < 3 {
		return false
	}
	return ast[0].Type() == "data" && ast[1].Value() == "." && ast[2].Type() == "data"
}
func itlookslikeastring(this lexer.Ast) bool {
	s := this.Value()
	if len(s) < 2 {
		return false
	}
	return s[0] == '"' && s[len(s)-1] == '"'
}
func (context function) eval(scanner *bufio.Scanner, usages interpreter) (keyword string, line int, err error) {
	if usages.stuff == nil {
		usages.stuff = make(variables)
	}
	if usages.funcs == nil {
		usages.funcs = make(functions)
	}
	if usages.packaged == nil {
		usages.packaged = make(map[string]functions)
	}
	if usages.structs == nil {
			usages.structs = make(map[string]structure)
	}
	for i := range context.args {
		usages.stuff[context.argnames[i]] = new(variable)
		*usages.stuff[context.argnames[i]] = context.args[i]
	}
	var ast []lexer.Ast
	for scanner.Scan() {
		filesline := scanner.Text()
		line++
		ast, err = lexer.Lex(filesline)
		if err != nil {
			return
		}
		if ast == nil {
		} else if len(ast) < 2 && ast[0].Value() != "return" && ast[0].Value() != "break" && ast[0].Value() != "continue" {
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
				err = fmt.Errorf("cannot return nil from function with %v", context.returns)
				return
			}
			keyword = "return"
			var vars []variable
			vars, err = usages.tuple(ast[1:])
			if err != nil {
				return
			}
			if len(vars) != len(context.returns) {
				err = fmt.Errorf("expected %v got %v", vars, context.returns)
			}
			for i := range context.returns {
				err = context.returns[i].assign(vars[i])
				if err != nil {
					return
				}
			}
			return
		} else if ast[0].Value() == "structure" {
			structpath := ast[1].Value()
			if !itlookslikeastring(ast[1]) {
				err = fmt.Errorf("%v does not look like a string", structpath)
				return
			}
			structpath = structpath[1 : len(structpath)-1]

			f, err2 := os.Open(structpath)
			if err2 != nil {
				err = err2
			}
			liner := bufio.NewScanner(f)
			splitpath := strings.Split(structpath,"/")
			structname := splitpath[len(splitpath)-1]
			usages.structs[structname]= structure{vars:make(map[string]string)}
			for liner.Scan() {
				ast, err2 := lexer.Lex(liner.Text())
				if err2 != nil {
					err = err2
					return
				}
				split := lexer.SplitFunc(ast,func(a lexer.Ast) bool {return a.Value() == ","})
				var line_T string
				tt := split[len(split)-1][1:]
				for i := range tt {
					line_T+=tt[i].Value()
				}
				split[len(split)-1] = split[len(split)-1][:1]
				for i := range split {
					if len(split[i]) > 1 {
						err = errors.Join(err,fmt.Errorf("invalid number of tokens in declaration %v",split[i]))
					} else if split[i][0].Type() != "data" {
						err = errors.Join(err,fmt.Errorf("invalid token in declaration %v",split[i]))
					}
					usages.structs[structname].vars[split[i][0].Value()] = line_T
				}
					
			}
			fmt.Println(usages.structs[structname],structname)
			if err != nil {
				return
			}
		} else if ast[0].Value() == "include" {
			packagepath := ast[1].Value()
			if !itlookslikeastring(ast[1]) {
				err = fmt.Errorf("%v does not look like a string", packagepath)
				return
			}
			packagepath = packagepath[1 : len(packagepath)-1]
			packagesplit := strings.Split(packagepath, "/")
			packagename := packagesplit[len(packagesplit)-1]
			funcnames, err2 := os.ReadDir(packagepath)
			if err2 != nil {
				err = err2
				return
			}
			packagename = strings.ReplaceAll(packagename, " ", "")
			usages.packaged[packagename] = make(functions)
			var collective error
			for _, e := range funcnames {
				if e.IsDir() {
					collective = errors.Join(collective, fmt.Errorf("cant have subdir in package %v", e.Name()))
					continue
				}
				usages.packaged[packagename][e.Name()], err = usages.newfunc(packagepath, e.Name())
				collective = errors.Join(collective, err)
			}
			if collective != nil {
				err = collective
				return
			}
		} else if ast[0].Value() == "find" {
			namestring := ast[1].Value()
			if !itlookslikeastring(ast[1]) {
				err = fmt.Errorf("%v does not look like a string", namestring)
				return
			}
			name := namestring[1 : len(namestring)-1]
			var current string
			current, err = os.Getwd()
			usages.funcs[name], err = usages.newfunc(current, name)
			if err != nil {
				return
			}
		} else if ast[0].Value() == "print" && ast[1].Value() == "()" {
			printed, err := usages.tuple(ast[1].Inner())
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%v\n", printed)
			}
		} else if ast[0].Value() == "the" {
			ast = ast[1:]
			n := slices.IndexFunc(ast, func(ast lexer.Ast) bool { return ast.Value() == "=" })
			if n == -1 {
				var collective error
				var d int
				for n = range ast {
					if ast[n].Type() == "data" || ast[n].Value() == "[]" {
						d++
					} else if ast[n].Value() == "," {
						d = 0
					} else {
						collective = errors.Join(collective, fmt.Errorf("invalid token %s", ast[n].Value()))
					}
					if d == 2 {
						break
					}
				}
				if collective != nil {
					err = collective
					return
				}
				assignedtolist := lexer.Split(ast[:n], lexer.Looker{Value: ","})
				var all variable
				all, err = newVariableOfType(ast[n:])
				if err != nil {
					return
				}
				for i := range assignedtolist {
					collective = errors.Join(collective, usages.stuff.New(assignedtolist[i][0], all))

				}
				if collective != nil {
					err = collective
					return
				}
			} else {
				assignedtolist := lexer.Split(ast[:n], lexer.Looker{Value: ","})
				if len(assignedtolist) == 1 {
					var val variable
					val, err = usages.valueof(ast[n+1:])
					if err != nil {
						return
					}
					err = usages.stuff.New(ast[0], val)
				} else {
					var vals []variable

					vals, err = usages.tuple(ast[n+1:])
					if err != nil {
						return
					}
					if len(vals) != len(assignedtolist) {
						err = fmt.Errorf("%v could not satisfy %v", assignedtolist, ast[n+1])
						return
					}
					for i := range assignedtolist {
						err = errors.Join(err, usages.stuff.New(assignedtolist[i][0], vals[i]))

					}
				}
			}
			if err != nil {
				fmt.Println(ast[:n], ":", err)
			}
		} else if itlookslikeassignment(ast) {
			n := slices.IndexFunc(ast, func(ast lexer.Ast) bool { return ast.Value() == "=" })
			if n >= len(ast)-1 {
				err = fmt.Errorf("%v does nothing", ast)
			}
			varnames := lexer.SplitFunc(ast[:n], func(a lexer.Ast) bool { return a.Value() == "," })
			var v []variable
			if len(varnames) == 1 {
				v = make([]variable, 1)
				v[0], err = usages.valueof(ast[n+1:])
			} else {
				v, err = usages.tuple(ast[n+1:])
			}
			if err != nil {
				return
			}
			if len(varnames) != len(v) {
				err = fmt.Errorf("%v does not satisfy %v", v, varnames)
				return
			}
			var collective error
			for i := range varnames {
				if len(varnames[i]) == 1 {
					place, ok := usages.stuff[varnames[i][0].Value()]
					if ok {
						err = (*place).assign(v[i])
					} else {
						err = errors.New(varnames[i][0].Value() + ":not found ")
					}
					collective = errors.Join(collective, err)
				} else {

					name := varnames[i][0].Value()
					varnames[i] = varnames[i][1:]
					var index []variable
					var val variable
					for j := range varnames[i] {
						if varnames[i][j].Value() != "[]" {
							err = fmt.Errorf("stray %v", varnames[i][j])
							return
						}
						val, err = usages.valueof(varnames[i][j].Inner())
						if err != nil {
							return
						}
						index = append(index, val)
					}
					if err != nil {
						fmt.Println(err)
						continue
					}
					place, ok := usages.stuff[name]
					if ok {
						err = (*place).setindex(index, v[i])
					} else {
						err = errors.New(name + ":not found ")
					}
				}
				collective = errors.Join(collective, err)

			}
			if collective != nil {
				err = collective
				return
			}
		} else if ast[0].Value() == "if" && ast[len(ast)-1].Value() == "then" {
			if len(ast) == 2 {
				err = errors.New("if but no condition")
				return
			}
			var condition int
			var v variable
			v, err = usages.expr(ast[1 : len(ast)-1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			condition, err = asint(v)
			if err != nil {
				fmt.Println(err)
				continue
			}
			var ifLambda, elseLambda string
			var lambda *string = &ifLambda
			var collective error
			var sp int = 1
			for scanner.Scan() {
				filesline = scanner.Text()
				ast, err := lexer.Lex(filesline)
				if ast != nil {
					if ast[0].Value() == "end" && ast[len(ast)-1].Value() == "if" {
						sp--
					} else if ast[0].Value() == "if" && ast[len(ast)-1].Value() == "then" {
						sp++
					} else if ast[0].Value() == "else" && sp == 1 {
						lambda = &elseLambda
						continue
					}
				}
				collective = errors.Join(collective, err)
				if sp == 0 {
					break
				}
				*lambda += filesline + "\n"

			}
			if collective != nil {
				err = collective
				return
			}
			if condition >= 1 {
				keyword, line, err = context.eval(bufio.NewScanner(strings.NewReader(ifLambda)), usages.refer())
				if keyword == "return" {
					return
				}
				if err != nil {
					return
				}
			} else {
				keyword, line, err = context.eval(bufio.NewScanner(strings.NewReader(elseLambda)), usages.refer())
				if keyword == "return" {
					return
				}
				if err != nil {
					return
				}
			}
		} else if ast[0].Value() == "for" && ast[len(ast)-1].Value() == "loop" {
			expression := ast[1 : len(ast)-1]
			var lambda string
			var sp int = 1
			var collective error
			var lambdast []lexer.Ast
			for scanner.Scan() {
				filesline = scanner.Text()
				lambdast, err = lexer.Lex(filesline)
				if lambdast != nil && len(lambdast) >= 2 {
					if lambdast[0].Value() == "end" && lambdast[len(lambdast)-1].Value() == "loop" {
						sp--
					} else if lambdast[0].Value() == "for" && lambdast[len(lambdast)-1].Value() == "loop" {
						sp++
					}
				}
				collective = errors.Join(collective, err)
				if sp == 0 {
					break
				}
				lambda += filesline + "\n"

			}
			if !(lambdast[0].Value() == "end" && lambdast[len(lambdast)-1].Value() == "loop") {
				err = fmt.Errorf("loop with no end")
				return

			}
			var returnedkeyword string
			var condition int
			var f function
			var v variable
			for {
				v, err = usages.expr(expression)
				if err != nil {
					return
				}

				condition, err = asint(v)
				if err != nil {
					return
				}
				returnedkeyword, line, err = f.eval(bufio.NewScanner(strings.NewReader(lambda)), usages.refer())
				switch returnedkeyword {
				case "break":
					break
				case "continue":

				}
				if err != nil {
					return
				}
				if condition <= 0 {
					break
				}
			}
		} else if itlookslikeafunctioncall(ast) {
			var returned []variable
			returned, err = usages.call(ast)
			if err != nil {
				return
			}
			if len(returned) != 0 {
				err = fmt.Errorf("%v return variables ignored", filesline)
				return
			}
		} else {
			fmt.Println("error", filesline, "looks like gibberish")
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
	var line int
	keyword, line, err = f.eval(bufio.NewScanner(file), interpreter{})
	if err != nil {
		fmt.Printf("main:%v:%v\n", line, err)
	}
	if keyword != "return" && keyword != "" {
		fmt.Println("error stray", keyword)
	}

}
