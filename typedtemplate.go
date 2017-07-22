package typedtemplate

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template/parse"
)

func fprintTypedTemplate(b io.Writer, name string, m map[string]interface{}) {
	structName := fmt.Sprintf("%sData", name)
	fmt.Fprintf(b, "type %sInterface interface {\n", name)
	fmt.Fprintf(b, "  Execute(io.Writer, %s) error\n", structName)
	fmt.Fprintf(b, "}\n")
	fprintStruct(b, structName, m)
}

func fprintStruct(b io.Writer, name string, m map[string]interface{}) {
	fmt.Fprintf(b, "type %s struct {\n", name)
	fprintStructInner(b, m, 1)
	fmt.Fprintf(b, "}\n")
}

func fprintStructInner(b io.Writer, m map[string]interface{}, level int) {

	nextLevel := level + 1
	pad := strings.Repeat(" ", level*2)
	keys := make([]string, len(m))
	i := 0
	for key, _ := range m {
		keys[i] = key
		i++
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		v1 := v.(map[string]interface{})
		if len(v1) > 0 {
			if strings.HasSuffix(k, "[]") {
				k2 := k[:len(k)-2]
				fmt.Fprintf(b, "%s%s []struct {\n", pad, k2)
			} else {
				fmt.Fprintf(b, "%s%s struct {\n", pad, k)
			}
			fprintStructInner(b, v1, nextLevel)
			fmt.Fprintf(b, "%s}\n", pad)
		} else {
			if strings.HasSuffix(k, "[]") {
				k2 := k[:len(k)-2]
				fmt.Fprintf(b, "%s%s []string\n", pad, k2)
			} else {
				fmt.Fprintf(b, "%s%s string\n", pad, k)
			}
		}
	}
}

func stripDollar(s [][]string) [][]string {
	b := make([][]string, len(s))
	for i, st := range s {
		if st[0] == "$" {
			b[i] = st[1:]
		} else {
			b[i] = st
		}
	}
	return b
}

type InterfaceDef map[string]interface{}

func interfaceTree(s [][]string) InterfaceDef {
	b := make(map[string]interface{})
	var lastMap map[string]interface{}
	for _, sa := range s {
		lastMap = b
		for _, sb := range sa {
			m, ok := lastMap[sb]
			if ok {
				lastMap = m.(map[string]interface{})
			} else {
				lastMap[sb] = make(map[string]interface{})
				lastMap = lastMap[sb].(map[string]interface{})
			}
		}
	}
	return b
}

func cannonicalVarNames(s [][]string) []string {
	b := make([]string, len(s))
	for i, st := range s {
		b[i] = fmt.Sprintf(".%s", strings.Join(st, "."))
	}
	return b
}

func extractVariables(n parse.Node) [][]string {

	b := [][]string{}

	switch n := n.(type) {
	case nil:
		return b
	case *parse.FieldNode:
		b = append(b, n.Ident)
	case *parse.ActionNode:
		vars := extractVariables(n.Pipe)
		b = append(b, vars...)
	case *parse.IfNode:
		b = append(b, extractVariables(n.Pipe)...)
		b = append(b, extractVariables(n.List)...)
		b = append(b, extractVariables(n.ElseList)...)
	case *parse.RangeNode:
		var prefix []string
		if n.Pipe != nil {
			t := extractVariables(n.Pipe)
			prefix = t[0]
		}
		prefix[len(prefix)-1] = fmt.Sprintf("%s[]", prefix[len(prefix)-1])
		vars := extractVariables(n.List)

		//fmt.Printf("%#v\n", vars)
		//if string(n.Ident[0][0]) == "$" {
		//	fmt.Printf("%#v\n", n.Ident)
		//}

		if len(vars) == 0 {
			b = append(b, prefix)
		} else {
			prefixedVars := make([][]string, 0, len(vars))
			for _, v := range vars {
				if v[0] == "$" {
					prefixedVars = append(prefixedVars, v)
				} else {
					if string(v[0][0]) != "$" {
						prefixedVars = append(prefixedVars, append(prefix, v...))
					}
				}
			}
			if len(prefixedVars) == 0 {
				b = append(b, prefix)
			} else {
				b = append(b, prefixedVars...)
			}
		}
	case *parse.WithNode:
		var prefix []string
		if n.Pipe != nil {
			t := extractVariables(n.Pipe)
			prefix = t[0]
		}
		vars := extractVariables(n.List)
		prefixedVars := make([][]string, len(vars))
		for i, v := range vars {
			if v[0] == "$" {
				prefixedVars[i] = v
			} else {
				prefixedVars[i] = append(prefix, v...)
			}
		}
		b = append(b, prefixedVars...)
	case *parse.ListNode:
		for _, node := range n.Nodes {
			b = append(b, extractVariables(node)...)
		}
	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			b = append(b, extractVariables(cmd)...)
		}
	case *parse.VariableNode:
		b = append(b, n.Ident)
	case *parse.CommandNode:
		for _, arg := range n.Args {
			b = append(b, extractVariables(arg)...)
		}
	}
	return b
}
