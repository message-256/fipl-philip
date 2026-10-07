package lexer
import (
	"strings"
	"github.com/message-256/escpaper2"
	"slices"
	"errors"
	"fmt"
)
func delimrecursive (s string ,delimitera,delimiterb rune) (int,error){
	var next int
	var sp int
	for next = 0; next<len(s); next++ {
		if s[next] == '#' {
			break
		}
		if s[next] == '"' {
			if peek(s,next) == 0{
				return 0, errors.New("stray \"")
			}
			if peek(s,next) == '"' {
				next+=2
			} else {
				str := escpaper2.SubString(s[next+1:],'"')
				if str[len(str)-1] != '"' {
					return 0,errors.New("a string with no end")
				}
				next+=len(str)+1
			}
		}
		if rune(s[next]) == delimitera {
			sp++
		} else if rune(s[next]) == delimiterb {
			sp--
		}
		if sp == 0 {
			break
		}
	}
	if sp > 0 {
		return 0,errors.New("stray " + string(delimitera))
	}
	return next,nil
	
}
func notbracketSpecialCharacter(c rune) bool {
	specials := "!=^%&*+-?/<>,:;$`@|.#"
	for i := range specials {
		if rune(specials[i]) == c {
			return true
		}
	}
	return false
}
func someKindOfBracket(c rune) bool {
	brackets := "{}[]()\"'`"
	for i := range brackets {
		if rune(brackets[i]) == c {
			return true
		}
	}
	return false
}
func specialcharacterf(c rune) bool {
	return notbracketSpecialCharacter(c) || someKindOfBracket(c)
}
type Ast interface {
	Type() string
	Value() string 
	Inner() []Ast

}

type perens struct {
	name string
	inner []Ast
}
type arithmatic struct {
	name string
}
type data struct {
	name string
}
func (d data) Type() string {
	return "data"
}
func (a arithmatic) Type() string {
	return "arithmatic"
}
func (p perens)Type() string {
	return "perenthesis"	
}
func (d data) Value() string {
	return d.name
}
func (a arithmatic) Value() string {
	return a.name
}
func (p perens) Value() string {
	return p.name	
}
func (d data) Inner() []Ast {
	return nil
}
func (a arithmatic) Inner() []Ast {
	return nil
}
func (p perens) Inner() []Ast {
	return p.inner	
}
func peek(this string,place int) rune {
	if place+1 >= len(this)-1{
		return 0
	} else {
		return rune(this[place+1])
	}
}
func (p perens)String() string {
	return fmt.Sprintf("name:'perenthesis',value:'%s',inner:'%v'",p.Value(),p.Inner())
}
func (d data)String() string {
	return fmt.Sprintf("name:'data',value:'%s'",d.Value())
}
func (a arithmatic)String() string {
	return fmt.Sprintf("name:'arithmatic',value:'%s'",a.Value())
}
func Lex(this string) ([]Ast,error){	
	if this == "" {
		return nil,nil
	}
	var returned []Ast
	var next = map[rune]rune {'{':'}',
		 '(':')',
		 '[':']',
	}
	for this != "" {
		i := strings.IndexFunc(this,specialcharacterf)
		
		if i == -1 {
			namesquestionmark := this
			namesquestionmark = strings.ReplaceAll(namesquestionmark,"\t","")
			namesSplitwithSpaces := strings.Split(namesquestionmark," ")
			namesclean := slices.DeleteFunc(namesSplitwithSpaces,func(s string)bool {return s == ""})
			for i := range namesclean {
				returned = append(returned,data{name:namesclean[i]})
			}
			return returned,nil
		}
		if this[i] == '#' {
			return returned,nil
		}
		if strings.ContainsFunc(this[:i],func(c rune) bool {return !(c == ' ')}){
			namesquestionmark := this[:i]
			namesquestionmark = strings.ReplaceAll(namesquestionmark,"\t","")
			namesSplitwithSpaces := strings.Split(namesquestionmark," ")
			namesclean := slices.DeleteFunc(namesSplitwithSpaces,func(s string)bool {return s == ""})
			for i := range namesclean {
				returned = append(returned,data{name:namesclean[i]})
			}
		}
		this = this[i:]
		
		if notbracketSpecialCharacter(rune(this[0])){
			i := strings.IndexFunc(this,func(c rune ) bool {return !notbracketSpecialCharacter(c)})
			returned = append(returned,arithmatic{name:string(this[:i])})
			this = this[i:]	
		}else if this[0] == '"'{
			if peek(this,0) == '"' {
				returned = append(returned,data{name:"\"\""})
			} else {
				str := escpaper2.SubString(this[1:],'"')
				if str[len(str)-1] != '"' {
					return returned,errors.New(fmt.Sprintf("%s string with no end",str))
				}
				returned = append(returned,data{name:"\"" + str})
				this = this[len(str)+1:]
			}
			
		} else if someKindOfBracket(rune(this[0])) {
			var inner []Ast
			var i int
			var err error
			i ,err = delimrecursive(this,rune(this[0]),next[rune(this[0])])
			if err != nil {
				return returned,err
			}
			inner,err = Lex(this[1:i])
			if err != nil {
				return returned,err
			}
			returned = append(returned,perens{name:string(this[0])+string(next[rune(this[0])]),inner:inner})
			this = this[i+1:]
		}
	}	
	return returned,nil

}
const (
	typename = iota
	value
)
type Looker struct {
	Type string
	Value string
}

func YouSee(a []Ast,b []Looker) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if (a[i].Type() != b[i].Type && b[i].Type != "") || (a[i].Value() != b[i].Value && b[i].Type == "") {
			fmt.Printf("a == %v,b == %v\n",a[i],b[i])

			return false
		}
	}
	return true
}
func SplitFunc(a []Ast, splitter func(Ast)bool) [][]Ast {
	var returned [][]Ast
	for len(a) != 0 {
		next := slices.IndexFunc(a,splitter)
		if next == -1 {
			return append(returned,a)
		}
		returned = append(returned,a[:next])
		a = a[next+1:]
	
	}
	return returned

}
func Split(a []Ast,b Looker) [][]Ast {
	var indexer func(Ast) bool
	if (b.Value == "" && b.Type == "") || len(a) == 0 {
		return nil
	}
	if b.Type == "" {
		indexer = func(ast Ast) bool {
			return ast.Value() == b.Value
		}
	} else if b.Value == "" {
		indexer = func(ast Ast) bool {
			return ast.Type() == b.Type
		}	
	} else {
		indexer = func(ast Ast) bool {
			return ast.Type() == b.Type && ast.Value() == b.Value
		}
	}
	return SplitFunc(a,indexer)
}