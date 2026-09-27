package lexer_test
import (
	"philip/lexer"
	"testing"
	"fmt"
)
func TestLexer(tester *testing.T) {
	
	var tests = []string { "test()","1+2","if i == 1 then","the b = \"test\"","test(\"hello\")","print(1+1,abc)","(1==1)","include "hello/stuff"","print(,hello)"}
	for i := range tests {
		fmt.Println("giving:",tests[i])
		ast,err := lexer.Lex(tests[i])
		fmt.Printf("ast:%v,err:%v\n",ast,err)

		if lexer.YouSee(ast,[]lexer.Looker{{Type:"data"},{Type:"perenthesis",Value:"()"}}) {
			inner := ast[1].Inner()
			split := lexer.Split(inner,lexer.Looker{Value:","})
			fmt.Printf("\nfound the function %v with args %v\n\n",ast,split)
		}
	}
}
