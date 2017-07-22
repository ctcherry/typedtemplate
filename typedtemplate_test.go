package typedtemplate

import (
	"testing"

	"bytes"
	"fmt"
	"html/template"
)

func TestSomething(*testing.T) {

}

func ExampleOutputGoTypes() {
	const letter = `
Dear {{.Person.Name}},
{{if .Person.Attended}}
It was a pleasure to see you at the wedding.{{else}}
It is a shame you couldn't make it to the wedding.{{end}}
{{with .Person}}
Thank you for the lovely {{.Gift.Item}}. I Hope it didnt cost more than {{.Gift.Price}}
	{{with $.OtherPerson}}
		Did you meet my friend {{.Name}}? He would have been interested in meeting the Great {{$.Person.FancyName}}
	{{end}}
{{end}}

Some of the other Gifts I got were:
{{range .OtherGifts}}
- {{.Item}} cost {{.Price}}
{{end}}

Some of the other People at the party who weren't invite:
{{range .NotInvitedPeople}}
- {{.}}
{{end}}

I think this people would make good pairs
{{range $index, $name := .PairsOfPeople}}
- {{$index}} with {{$name}}
{{end}}
Best wishes,
Josie
`
	tmpl, err := template.New("letter").Parse(letter)

	if err != nil {
		fmt.Printf("%#v\n", err)
	} else {
		ifaceTree := interfaceTree(stripDollar(extractVariables(tmpl.Tree.Root)))
		b := &bytes.Buffer{}
		fprintTypedTemplate(b, "Letter", ifaceTree)
		fmt.Println(b)
	}

	// Output:
	// type LetterInterface interface {
	//   Execute(io.Writer, LetterData) error
	// }
	// type LetterData struct {
	//   NotInvitedPeople []string
	//   OtherGifts []struct {
	//     Item string
	//     Price string
	//   }
	//   OtherPerson struct {
	//     Name string
	//   }
	//   PairsOfPeople []string
	//   Person struct {
	//     Attended string
	//     FancyName string
	//     Gift struct {
	//       Item string
	//       Price string
	//     }
	//     Name string
	//   }
	// }

}
