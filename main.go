package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"strconv"

	//"go/types"
	"golang.org/x/tools/go/loader"
	"log"
)

func main() {
	testData := "./test"

	ctx := &Context{}
	ctx.Load(testData)
	errs := ctx.Process()
	for _, err := range errs {
		fmt.Println("%v, %v", err.VarName, err.Line)
	}
}

type Context struct {
	cwd string
	loader.Config
}

func (ctx *Context) Load(args string) {
	ctx.Config.Import(args)
}

func (ctx *Context) Process() DivisionZeroErrors {
	prog, err := ctx.Config.Load()
	if err != nil {
		log.Fatalf("cannot load packages: %s", err)
	}

	var allDivisionZero DivisionZeroErrors
	for _, pkg := range prog.Imported {
		divisionZero := doPackage(prog, pkg)
		allDivisionZero = append(allDivisionZero, divisionZero...)
	}

	return allDivisionZero
}

func doPackage(prog *loader.Program, pkg *loader.PackageInfo) DivisionZeroErrors {

	//checked := make(map[string]bool)
	varValue := make(map[string]string)
	errs := make([]*DivisionZeroError, 0)

	for _, file := range pkg.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			ttype := reflect.TypeOf(node)
			if ttype != nil {
				switch stmt := node.(type) {
				case *ast.AssignStmt:
					if stmt.Tok.String() == "=" || stmt.Tok.String() == ":=" {
						for i, lhs := range stmt.Lhs {
							id, ok := lhs.(*ast.Ident)
							if ok {
								val, vok := stmt.Rhs[i].(*ast.BasicLit)
								if vok {
									varValue[id.Name] = val.Value
								}
							}
						}
					} else if stmt.Tok.String() == "/" || stmt.Tok.String() == "/=" {
						a, _ := stmt.Lhs[0].(*ast.Ident)
						b, _ := stmt.Rhs[0].(*ast.Ident)
						beDivided := varValue[b.Name]
						beDividedVal, _ := strconv.ParseFloat(beDivided, 10)
						if beDividedVal == 0.0 {

							errs = append(errs, &DivisionZeroError{
								VarName: a.Name,
								Line:    fmt.Sprintf("%v", a.NamePos),
							})
						}
					}

				default:
				}
				return true
			}

			return false
		})
	}

	for k, v := range varValue {
		fmt.Printf("%v: %v", k, v)
		fmt.Println()
	}
	return errs
}

type DivisionZeroError struct {
	VarName string
	Line    string
}

type DivisionZeroErrors []*DivisionZeroError
