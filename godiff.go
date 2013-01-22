package main

import (
    "github.com/daviddengcn/go-villa"
    "bytes"
    "fmt"
    "github.com/daviddengcn/go-algs/ed"
    "go/ast"
    "go/parser"
    "go/printer"
    "go/token"
    "sort"
    "strings"
//    "log"
    "os"
)

func cat(a, sep, b string) string {
	if len(a) > 0 && len(b) > 0 {
		return a + sep + b
	} // if

	return a + b
}

func max(a, b int) int {
	if a > b {
		return a
	} // if

	return b
}

func MakeDiffMatrix(lenA, lenB int, diffF func(iA, iB int) int) villa.IntMatrix {
    mat := villa.NewIntMatrix(lenA, lenB)
    
    for iA := 0; iA < lenA; iA ++ {
        for iB := 0; iB < lenB; iB ++ {
            mat[iA][iB] = diffF(iA, iB)
        } // for iB
    } // for iA
    
    return mat
}

func GreedyMatch(diffMat villa.IntMatrix, delCost, insCost func(int) int) (cost int, matA, matB []int) {
    mat := diffMat.Clone()
    nRows, nCols := mat.Rows(), mat.Cols()
    // mx is a number greater or equal to all mat elements (need not be the exact maximum)
    mx := 0
    for r := range mat {
        for c := 0; c < nCols; c ++ {
            mat[r][c] -= delCost(r) + insCost(c)
            if mat[r][c] > mx {
                mx = mat[r][c]
            } // if
        } // for c
    } // for r
    
    matA, matB = make([]int, nRows), make([]int, nCols)
    villa.IntSlice(matA).Fill(0, nRows, -1)
    villa.IntSlice(matB).Fill(0, nCols, -1)
    
    for {
        mn := mx + 1
        selR, selC := -1, -1
        for r := range mat {
            if matA[r] >= 0 {
                continue
            } // if
            for c := 0; c < nCols; c ++ {
                if matB[c] >= 0 {
                    continue
                } // if
                
                if mat[r][c] < mn {
                    mn = mat[r][c]
                    selR, selC = r, c
                } // if
            } // for c
        } // for r
        
        if selR < 0 || mn >= 0 {
            break
        } // if
        
        matA[selR] = selC
        matB[selC] = selR
    } // for
    
    for c := range(matA) {
        if matA[c] < 0 {
            cost += delCost(c)
        } else {
            cost += diffMat[c][matA[c]]
        } // else
    } // for c
    for r := range(matB) {
        if matB[r] < 0 {
            cost += insCost(r)
        } // if
    } // for r
    
    return cost, matA, matB
}


const(
    DF_NONE = iota
    DF_TYPE
    DF_CONST
    DF_VAR
    DF_STRUCT
    DF_INTERFACE
    DF_FUNC
    DF_STAR
    DF_VAR_LINE
    DF_PAIR
    DF_NAMES
    DF_VALUES
    DF_BLOCK
    DF_RESULTS
)

var TYPE_NAMES []string = []string {
        "",
        "type",
        "const",
        "var",
        "struct",
        "interface",
        "func",
        "*",
        "",
        "",
        "",
        "",
        "",
        ""}

type DiffFragment interface {
    Type() int
    Weight() int
    calcDiff(that DiffFragment) int
    showDiff(that DiffFragment)
    // indent is the leading chars from the second line
    sourceLines(indent string) []string
    oneLine() string
}

type Fragment struct {
    tp int
    Parts []DiffFragment
}

func (f *Fragment) Type() int {
    return f.tp
}

func (f *Fragment) Weight() int {
    if f == nil {
        return 0
    } // if
    
    w := 0
    for _, p := range f.Parts {
        w += p.Weight()
    } // for p
    return w
}

func catLines(a []string, sep string, b []string) []string {
    if len(a) > 0 && len(b) > 0 {
        b[0] = cat(a[len(a) - 1], sep, b[0])
        a = a[0:len(a) - 1]
    } // if
    
    return append(a, b...)
}

func appendLines(a []string, sep string, b ...string) []string {
    if len(a) > 0 && len(b) > 0 {
        b[0] = cat(a[len(a) - 1], sep, b[0])
        a = a[0:len(a) - 1]
    } // if
    
    return append(a, b...)
}

func insertIndent(indent string, lines []string) []string {
    for i := range(lines) {
        lines[i] = indent + lines[i]
    } // for i
    
    return lines
}

func insertIndent2(indent string, lines []string) []string {
    for i := range(lines) {
        if i > 0 {
            lines[i] = indent + lines[i]
        } // if
    } // for i
    
    return lines
}

func (f *Fragment) oneLine() string {
    if f == nil {
        return ""
    } // if
    switch f.tp {
    } // switch
    lines := f.sourceLines("")
    if len(lines) == 0 {
        return ""
    } // if
    
    if len(lines) == 1 {
        return lines[0]
    } // if
    
    return lines[0] + " ... " + lines[len(lines) - 1]
}

func (f *Fragment) sourceLines(indent string) (lines []string) {
    if f == nil {
        return nil
    } // if
    switch f.tp {
        case DF_TYPE:
            lines = append(lines, TYPE_NAMES[f.tp])
            lines = catLines(lines, " ", f.Parts[0].sourceLines(indent))
            lines = catLines(lines, " ", f.Parts[1].sourceLines(indent))
        case DF_CONST:
            if len(f.Parts) == 1 {
                lines = append(lines, TYPE_NAMES[f.tp])
                lines = catLines(lines, " ", f.Parts[0].sourceLines(indent))
            } else {
                lines = append(lines, TYPE_NAMES[f.tp] + "(")
                for _, p := range f.Parts {
                    lines = append(lines, catLines([]string{indent + "    "}, "", p.sourceLines(indent + "    "))...)
                } // p
                lines = append(lines, indent + ")")
            } // else
        case DF_VAR:
            lines = append(lines, TYPE_NAMES[f.tp])
            lines = catLines(lines, " ", f.Parts[0].sourceLines(indent + "    "))
        case DF_VAR_LINE:
            lines = f.Parts[0].sourceLines(indent)
            lines = catLines(lines, " ", f.Parts[1].sourceLines(indent))
            lines = catLines(lines, " = ", f.Parts[2].sourceLines(indent))
        case DF_FUNC:
            lines = append(lines, TYPE_NAMES[f.tp])
            if f.Parts[0].(*Fragment) != nil {
                lines = catLines(catLines(lines, " (", f.Parts[0].sourceLines(indent + "    ")), "", []string{")"}) // recv
            } // if
            lines = catLines(lines, " ", f.Parts[1].sourceLines(indent + "    ")) // name
            lines = catLines(catLines(catLines(lines, "", []string{"("}), "", f.Parts[2].sourceLines(indent + "    ")), "", []string{")"}) // params
            lines = catLines(lines, " ", f.Parts[3].sourceLines(indent + "    ")) // returns
            lines = catLines(lines, " ", f.Parts[4].sourceLines(indent)) // body
        case DF_RESULTS:
            if len(f.Parts) > 0 {
                if len(f.Parts) > 1 || len(f.Parts[0].(*Fragment).Parts[0].(*StringFrag).source) > 0 {
                    lines = append(lines, "(")
                } // if
                for i, p := range f.Parts {
                    if i > 0 {
                        lines = catLines(lines, "", []string{", "})
                    } // if
                    lines = catLines(lines, "", p.sourceLines(indent + "    "))
                } // for i, p
                if len(f.Parts) > 1 || len(f.Parts[0].(*Fragment).Parts[0].(*StringFrag).source) > 0 {
                    lines = catLines(lines, "", []string{")"})
                } // if
            } // if            
        case DF_BLOCK:
            lines = append(lines, "{")
            for _, p := range f.Parts {
                lines = append(lines, catLines([]string{indent + "    "}, "", p.sourceLines(indent + "    "))...)
            } // for p
            lines = append(lines, indent + "}")
        case DF_STRUCT, DF_INTERFACE:
            lines = append(lines, TYPE_NAMES[f.tp] + " {")
            for _, p := range f.Parts {
                lns := p.sourceLines(indent + "    ")
                if len(lns) > 0 {
                    lns[0] = indent + "    " + lns[0]
                    lines = append(lines, lns...)
                } // if
            } // for p
            lines = append(lines, indent + "}")
        case DF_STAR:
            lines = append(lines, TYPE_NAMES[f.tp])
            lines = catLines(lines, "", f.Parts[0].sourceLines(indent))
        case DF_PAIR:
            lines = catLines(f.Parts[0].sourceLines(indent), " ", f.Parts[1].sourceLines(indent))
        case DF_NAMES:
            s := ""
            for _, p := range f.Parts {
                s = cat(s, ", ", p.sourceLines(indent + "    ")[0])
            } // for p
            lines = append(lines, s)
        case DF_VALUES:
            for _, p := range f.Parts {
                lines = catLines(lines, ", ", p.sourceLines(indent + "    "))
            } // for p
        case DF_NONE:
            for _, p := range f.Parts {
                lines = append(lines, p.sourceLines(indent)...)
            } // for p
        default:
            lines = []string{"TYPE: " + TYPE_NAMES[f.Type()]}
            for _, p := range f.Parts {
                lines = append(lines, p.sourceLines(indent + "    ")...)
            } // for p
    }
    return lines
}

func (f *Fragment) calcDiff(that DiffFragment) int {
    if f == nil {
        return that.Weight()
    } // if
    
    switch g := that.(type) {
        case *Fragment:
            if g == nil || f.Type() != g.Type() {
                return f.Weight() + g.Weight()
            } // if
            
            return ed.EditDistanceF(len(f.Parts), len(g.Parts), func(iA, iB int) int {
                return f.Parts[iA].calcDiff(g.Parts[iB])
            }, func(iA int) int {
                return f.Parts[iA].Weight()
            }, func(iB int) int {
                return g.Parts[iB].Weight()
            })
    }
    return f.Weight() + that.Weight()
}

func (f *Fragment) showDiff(that DiffFragment) {
    DiffLines(f.sourceLines(""), that.sourceLines(""), `%s`)
}

type StringFrag struct {
    weight int
    source string
}

func newStringFrag(source string, weight int) *StringFrag {
    return &StringFrag{weight: weight, source: source}
}

func (sf *StringFrag) Type() int {
    return DF_NONE
}

func (sf *StringFrag) Weight() int {
    return sf.weight
}

func (sf *StringFrag) calcDiff(that DiffFragment) int {
    switch g := that.(type) {
        case *StringFrag:
            if len(sf.source) + len(g.source) == 0 {
                return 0
            } // if
            wt := sf.weight + g.weight
            return ed.String(sf.source, g.source)*wt/max(len(sf.source), len(g.source))
    } // switch
    
    return sf.Weight() + that.Weight()
}

func (sf *StringFrag) showDiff(that DiffFragment) {
    DiffLines(sf.sourceLines("    "), that.sourceLines("    "), `%s`)
}

func (sf *StringFrag) oneLine() string {
    if sf == nil {
        return ""
    } // if
    
    return sf.source
}

func (sf *StringFrag) sourceLines(indent string) []string {
    lines := strings.Split(sf.source, "\n")
    for i := range(lines) {
        if i > 0 {
            lines[i] = indent + lines[i]
        } // if
    } // for i
    
    return lines
}

const (
	TD_STRUCT = iota
	TD_INTERFACE
	TD_POINTER
	TD_ONELINE
)

func newNameTypes(fs *token.FileSet, fl *ast.FieldList) (dfs []DiffFragment) {
	for _, f := range fl.List {
    	if len(f.Names) > 0 {
    		for _, name := range f.Names {
    			dfs = append(dfs, &Fragment{tp: DF_PAIR, Parts: []DiffFragment{newStringFrag(name.String(), 100), newTypeDef(fs, f.Type)}})
    		} // for name
    	} else {
    		// embedding
    		dfs = append(dfs, &Fragment{tp: DF_PAIR, Parts: []DiffFragment{newStringFrag("", 50), newTypeDef(fs, f.Type)}})
    	} // else
	} // for f
    
    return dfs
}

func newTypeDef(fs *token.FileSet, def ast.Expr) DiffFragment {
    switch d := def.(type) {
        case *ast.StructType:
            return &Fragment{tp: DF_STRUCT, Parts: newNameTypes(fs, d.Fields)}
            
        case *ast.InterfaceType:
            return &Fragment{tp: DF_INTERFACE, Parts: newNameTypes(fs, d.Methods)}
            
        case *ast.StarExpr:
        	return &Fragment{tp: DF_STAR, Parts: []DiffFragment{newTypeDef(fs, d.X)}}
    } // switch
    
    var src bytes.Buffer
    (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(&src, fs, def)
    return &StringFrag{weight: 50, source: src.String()}
}

func newTypeStmtInfo(fs *token.FileSet, name string, def ast.Expr) *Fragment {
    var f Fragment
    
    f.tp = DF_TYPE
    f.Parts = []DiffFragment{
        newStringFrag(name, 100),
        newTypeDef(fs, def)}
    
    return &f
}

func newExpDef(fs *token.FileSet, def ast.Expr) DiffFragment {
//    ast.Print(fs, def)
    
	var src bytes.Buffer
	(&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(&src, fs, def)
	return &StringFrag{weight: 100, source: src.String()}
}

func newVarSpecs(fs *token.FileSet, specs []ast.Spec) (dfs []DiffFragment) {
	for _, spec := range specs {
        f := &Fragment{tp: DF_VAR_LINE}
        
        names := &Fragment{tp: DF_NAMES}
		sp := spec.(*ast.ValueSpec)
		for _, name := range sp.Names {
			names.Parts = append(names.Parts, &StringFrag{weight: 100, source: fmt.Sprint(name)})
		} // for name
        f.Parts = append(f.Parts, names)

        
		if sp.Type != nil {
            f.Parts = append(f.Parts, newTypeDef(fs, sp.Type))
		} else {
            f.Parts = append(f.Parts, (*Fragment)(nil))
        } // else

        values := &Fragment{tp: DF_VALUES}
		for _, v := range sp.Values {
			values.Parts = append(values.Parts, newExpDef(fs, v))
		} // for v
        f.Parts = append(f.Parts, values)

		dfs = append(dfs, f)
	} // for spec

	return dfs
}

func printToLines(fs *token.FileSet, node interface{}) []string {
    var src bytes.Buffer
    (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(&src, fs, node)
    return strings.Split(src.String(), "\n")
}

func nodeToLines(fs *token.FileSet, node interface{}) (lines []string) {
    switch nd := node.(type) {
        case *ast.IfStmt:
            lines = append(lines, "if")
            if nd.Init != nil {
                lines = appendLines(lines, " ", nodeToLines(fs, nd.Init)...)
                lines = appendLines(lines, "", ";")
            } // if
            
            lines = catLines(lines, " ", nodeToLines(fs, nd.Cond))
            lines = catLines(lines, " ", []string{"{"})
            lines = append(lines, insertIndent("    ", blockToLines(fs, nd.Body))...)
            lines = append(lines, "}")
            if nd.Else != nil {
                //ast.Print(fs, st.Else)
                lines = catLines(lines, "", []string{" else "})
                lines = catLines(lines, "", nodeToLines(fs, nd.Else))
            } // if
        case *ast.AssignStmt:
            for _, exp := range nd.Lhs {
                lines = catLines(lines, ", ", nodeToLines(fs, exp))
            } // for i
            
            lines = catLines(lines, "", []string{" " + nd.Tok.String() + " "})
            
            for i, exp := range nd.Rhs {
                if i > 0 {
                    lines = catLines(lines, "", []string{", "})
                } // if
                lines = catLines(lines, "", nodeToLines(fs, exp))
            } // for i
            
        case *ast.ForStmt:
            lines = append(lines, "for")
            if nd.Cond != nil {
                lns := []string{}
                if nd.Init != nil {
                    lns = catLines(lns, "; ", nodeToLines(fs, nd.Init))
                } // if
                lns = catLines(lns, "; ", nodeToLines(fs, nd.Cond))
                if nd.Post != nil {
                    lns = catLines(lns, "; ", nodeToLines(fs, nd.Post))
                } // if
                
                lines = catLines(lines, " ", lns)
            } // if
            lines = catLines(lines, "", []string{" {"})
            lines = append(lines, insertIndent("    ", blockToLines(fs, nd.Body))...)
            lines = append(lines, "}")
        case *ast.RangeStmt:
            lines = append(lines, "for")
            lines = catLines(lines, " ", nodeToLines(fs, nd.Key))
            if nd.Value != nil {
                lines = catLines(lines, ", ", nodeToLines(fs, nd.Value))
            } // if
            lines = catLines(lines, "", []string{" " + nd.Tok.String() + " "})
            lines = catLines(lines, "", []string{" range"})
            lines = catLines(lines, " ", printToLines(fs, nd.X))
            lines = catLines(lines, "", []string{" {"})
            lines = append(lines, insertIndent("    ", blockToLines(fs, nd.Body))...)
            lines = append(lines, "}")

        case *ast.BlockStmt:
            lines = append(lines, "{")
            lines = append(lines, insertIndent("    ", blockToLines(fs, nd))...)
            lines = append(lines, "}")
        
        case *ast.ReturnStmt:
            lines = append(lines, "return")
            if nd.Results != nil {
                for i, e := range nd.Results {
                    if i == 0 {
                        lines = appendLines(lines, " ", nodeToLines(fs, e)...)
                    } else {
                        lines = appendLines(lines, ", ", nodeToLines(fs, e)...)
                    } // else
                } // for i, e
            } // if
            
        case *ast.DeferStmt:
            lines = append(lines, "defer")
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Call)...)
            
        case *ast.GoStmt:
            lines = append(lines, "go")
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Call)...)
            
        case *ast.SendStmt:
            lines = append(lines, nodeToLines(fs, nd.Chan)...)
            lines = appendLines(lines, " ", "<-")
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Value)...)
            
        case *ast.EmptyStmt:
            
        case *ast.SwitchStmt:
            lines = append(lines, "switch")
            if nd.Init != nil {
                lines = appendLines(lines, " ", nodeToLines(fs, nd.Init)...)
                lines = appendLines(lines, "", ";")
            } // if
            if nd.Tag != nil {
                lines = appendLines(lines, " ", nodeToLines(fs, nd.Tag)...)
            } // if
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Body)...)
            
        case *ast.TypeSwitchStmt:
            lines = append(lines, "switch")
            if nd.Init != nil {
                lines = appendLines(lines, " ", nodeToLines(fs, nd.Init)...)
                lines = appendLines(lines, "", ";")
            } // if
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Assign)...)
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Body)...)
            
        case *ast.CompositeLit:
            if nd.Type != nil {
                lines = append(lines, printToLines(fs, nd.Type)...)
            } // if
            lines = appendLines(lines, "", "{")
            for i, el := range nd.Elts {
                if i > 0 {
                    lines = appendLines(lines, "", ", ")
                } // if
                lines = appendLines(lines, "", nodeToLines(fs, el)...)
            } // for i, el
            lines = appendLines(lines, "", "}")
        case *ast.UnaryExpr:
            lines = append(lines, nd.Op.String())
            lines = appendLines(lines, "", nodeToLines(fs, nd.X)...)
            
        case *ast.CallExpr:
            lines = append(lines, nodeToLines(fs, nd.Fun)...)
            lines = appendLines(lines, "", "(")
            for i, a := range nd.Args {
                if i > 0 {
                    lines = appendLines(lines, "", ", ")
                } // if
                lines = appendLines(lines, "", nodeToLines(fs, a)...)
            } // for i, el
            if nd.Ellipsis > 0 {
                lines = appendLines(lines, "", "...")
            } // if
            lines = appendLines(lines, "", ")")
        case *ast.KeyValueExpr:
            lines = append(lines, nodeToLines(fs, nd.Key)...)
            lines = appendLines(lines, ": ", nodeToLines(fs, nd.Value)...)
        case *ast.FuncLit:
            lines = nodeToLines(fs, nd.Type)
            lines = appendLines(lines, " ", nodeToLines(fs, nd.Body)...)
            
        case *ast.CaseClause:
            if nd.List == nil {
                lines = append(lines, "default:")
            } else {
                lines = append(lines, "case ")
                for i, e := range nd.List {
                    if i > 0 {
                        lines = appendLines(lines, "", ", ")
                    } // if
                    lines = appendLines(lines, "", nodeToLines(fs, e)...)
                } // for i
                lines = appendLines(lines, "", ":")
            } // else
            
            for _, st := range nd.Body {
                lines = append(lines, insertIndent("    ", nodeToLines(fs, st))...)
            } // for
            
        default:
            return printToLines(fs, nd)
    }
    
    return lines
}

func blockToLines(fs *token.FileSet, blk *ast.BlockStmt) (lines []string) {
    for _, s := range blk.List {
        lines = append(lines, nodeToLines(fs, s)...)
    } // for s
    
    return lines
}

func newBlockDecl(fs *token.FileSet, blk *ast.BlockStmt) (f *Fragment) {
    f = &Fragment{tp: DF_BLOCK}
    lines := blockToLines(fs, blk)
    for _, line := range(lines) {
        f.Parts = append(f.Parts, &StringFrag{weight: 100, source: line})
    } // for line
    
    return f
}

func newFuncDecl(fs *token.FileSet, d *ast.FuncDecl) (f* Fragment) {
    f = &Fragment{tp: DF_FUNC}
    
    // recv
    if d.Recv != nil {
        f.Parts = append(f.Parts, newNameTypes(fs, d.Recv)...)
    } else {
        f.Parts = append(f.Parts, (*Fragment)(nil))
    } // else
    
    // name
    f.Parts = append(f.Parts, &StringFrag{weight: 100, source: fmt.Sprint(d.Name)})

    //  params
    if d.Type.Params != nil {
        f.Parts = append(f.Parts, &Fragment{tp: DF_NAMES, Parts: newNameTypes(fs, d.Type.Params)})
    } else {
        f.Parts = append(f.Parts, (*Fragment)(nil))
    } // else
    
    // Results
    if d.Type.Results != nil {
        f.Parts = append(f.Parts, &Fragment{tp: DF_RESULTS, Parts: newNameTypes(fs, d.Type.Results)})
    } else {
        f.Parts = append(f.Parts, (*Fragment)(nil))
    } // else
    
    // body
    if d.Body != nil {
        f.Parts = append(f.Parts, newBlockDecl(fs, d.Body))
    } else {
        f.Parts = append(f.Parts, (*Fragment)(nil))
    } // else
	return f
}

type FileInfo struct {
	f     *ast.File
	fs    *token.FileSet
	types *Fragment
	vars  *Fragment
	funcs *Fragment
}

func (info *FileInfo) collect() {
    info.types = &Fragment{}
    info.vars = &Fragment{}
    info.funcs = &Fragment{}
    
	for _, decl := range info.f.Decls {
		switch d := decl.(type) {
    		case *ast.GenDecl:
    			switch d.Tok {
        			case token.TYPE:
        				for i := range d.Specs {
        					spec := d.Specs[i].(*ast.TypeSpec)
        					//ast.Print(info.fs, spec)
        					ti := newTypeStmtInfo(info.fs, spec.Name.String(), spec.Type)
        					info.types.Parts = append(info.types.Parts, ti)
        				} // for i
        			case token.CONST:
        				// fmt.Println(d)
        				//ast.Print(info.fs, d)
                        v := &Fragment{tp: DF_CONST, Parts: newVarSpecs(info.fs, d.Specs)}
                        info.vars.Parts = append(info.vars.Parts, v)
        			case token.VAR:
        				//ast.Print(info.fs, d)
        				vss := newVarSpecs(info.fs, d.Specs)
        				for _, vs := range vss {
        					info.vars.Parts = append(info.vars.Parts, &Fragment{tp: DF_VAR, Parts: []DiffFragment{vs}})
        				} // for spec
        			case token.IMPORT:
        					// ignore
        			default:
        				// Unknow
        				fmt.Println(d)
    			} // switch d.tok
    		case *ast.FuncDecl:
                //fmt.Printf("%#v\n", d)
                fd := newFuncDecl(info.fs, d)
                info.funcs.Parts = append(info.funcs.Parts, fd)
                //ast.Print(info.fs, d)
            default:
                fmt.Println(d)
		} // switch decl.(type)
	} // for decl
}

func Parse(fn string) (*FileInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fn, nil, 0)
	if err != nil {
		return nil, err
	} // if

	info := &FileInfo{f: f, fs: fset}
	info.collect()

	return info, nil
}

func ShowDelWholeLine(line string) {
	fmt.Println("===", line)
}
func ShowDelLine(line string) {
	fmt.Println("---", line)
}

func ShowDelLines(lines []string, gapLines int) {
    if len(lines) <= gapLines*2 + 1 {
        for _, line := range lines {
            ShowDelLine(line)
        } // for line
        return
    } // if
    
    for i, line := range(lines) {
        if i < gapLines || i >= len(lines)-gapLines {
            ShowDelLine(line)
        } // if
        if i == gapLines {
            ShowDelWholeLine("    ...")
        } // if
    } // for i
}

func ShowInsLine(line string) {
	fmt.Println("+++", line)
}

func ShowInsWholeLine(line string) {
	fmt.Println("###", line)
}

func ShowInsLines(lines []string, gapLines int) {
    if len(lines) <= gapLines*2 + 1 {
        for _, line := range lines {
            ShowInsLine(line)
        } // for line
        return
    } // if
    
    for i, line := range(lines) {
        if i < gapLines || i >= len(lines)-gapLines {
            ShowInsLine(line)
        } // if
        if i == gapLines {
            ShowInsWholeLine("    ...")
        } // if
    } // for i
}

func ShowDiffLine(orgLine, newLine string) {
	ShowDelLine(orgLine)
	ShowInsLine(newLine)
}

func DiffLinesNoOrder(orgLines, newLines []string, format string) {
	sort.Strings(orgLines)
	sort.Strings(newLines)

	//fmt.Println(orgLines)
	//fmt.Println(newLines)

	for i, j := 0, 0; i < len(orgLines) || j < len(newLines); {
		//fmt.Println(i, len(orgLines), j, len(newLines), orgLines[i], newLines[j])
		switch {
		case i >= len(orgLines), j < len(newLines) && orgLines[i] > newLines[j]:
			fmt.Println("+++", fmt.Sprintf(format, newLines[j]))
			j++
		case j >= len(newLines), i < len(orgLines) && orgLines[i] < newLines[j]:
			fmt.Println("---", fmt.Sprintf(format, orgLines[i]))
			i++
		case orgLines[i] == newLines[j]:
			i++
			j++
		}
	} // for i, j
}

type lineOutput struct {
    sameLines []string
}

func (lo *lineOutput) outputIns(line string) {
    lo.end()
    ShowInsLine(line)
}

func (lo *lineOutput) outputDel(line string) {
    lo.end()
    ShowDelLine(line)
}

func (lo *lineOutput) outputSame(line string) {
    lo.sameLines = append(lo.sameLines, line)
}

func (lo *lineOutput) end() {
    if len(lo.sameLines) > 0 {
        fmt.Println("   ", lo.sameLines[0])
        if len(lo.sameLines) > 2 {
            fmt.Println("   ", "    ...")
        }
        if len(lo.sameLines) > 1 {
            fmt.Println("   ", lo.sameLines[len(lo.sameLines) - 1])
        } // if
    } // if
    
    lo.sameLines = nil
}

func DiffLines(orgLines, newLines []string, format string) {
	if len(orgLines)+len(newLines) == 0 {
		return
	} // if
	_, matA, matB := ed.EditDistanceFFull(len(orgLines), len(newLines), func(iA, iB int) int {
		return diffOfStrings(orgLines[iA], newLines[iB], 2000)
	}, ed.ConstCost(1000), ed.ConstCost(1000))
	//fmt.Println(matA, matB)

	//fmt.Println(orgLines)
	//fmt.Println(newLines)

    var lo lineOutput

    for i, j := 0, 0; i < len(orgLines) || j < len(newLines); {
        //fmt.Println(i, len(orgLines), j, len(newLines), orgLines[i], newLines[j])
        switch {
            case j >= len(newLines) || i < len(orgLines) && matA[i] < 0:
                lo.outputDel(fmt.Sprintf(format, orgLines[i]))
            	i++
            case i >= len(orgLines) || j < len(newLines) && matB[j] < 0:
                lo.outputIns(fmt.Sprintf(format, newLines[j]))
            	j++
            default:
            	if strings.TrimSpace(orgLines[i]) != strings.TrimSpace(newLines[j]) {
                    lo.outputDel(fmt.Sprintf(format, orgLines[i]))
                    lo.outputIns(fmt.Sprintf(format, newLines[j]))
            	} else {
            		lo.outputSame(orgLines[i])
            	} // else
            	i++
            	j++
            }
    } // for i, j
    lo.end()
}

/*
   Diff Package
*/

func DiffPackages(orgInfo, newInfo *FileInfo) {
	fmt.Println("===== PACKAGES")
	defer fmt.Println("      PACKAGES =====")
	orgName := orgInfo.f.Name.String()
	newName := newInfo.f.Name.String()
	if orgName != newName {
		ShowDiffLine("package "+orgName, "package "+newName)
	} //  if
}

/*
   Diff Imports
*/

func extractImports(info *FileInfo) []string {
	imports := make([]string, 0, len(info.f.Imports))
	for _, imp := range info.f.Imports {
		imports = append(imports, imp.Path.Value)
	} // for imp

	return imports
}

func DiffImports(orgInfo, newInfo *FileInfo) {
	fmt.Println("===== IMPORTS")
	defer fmt.Println("      IMPORTS =====")
	orgImports := extractImports(orgInfo)
	newImports := extractImports(newInfo)
	//fmt.Println(orgImports)
	//fmt.Println(newImports)

	DiffLinesNoOrder(orgImports, newImports, `import %s`)
	//ast.Print(orgInfo.fs, orgInfo.f.Imports)
	//ast.Print(newInfo.fs, newInfo.f.Imports)
}

/*
   Diff Types
*/

func diffOfStrings(a, b string, mx int) int {
	if a == b {
		return 0
	} // if
	return ed.String(a, b) * mx / max(len(a), len(b))
}

func DiffTypes(orgInfo, newInfo *FileInfo) {
	fmt.Println("===== TYPES")
	defer fmt.Println("      TYPES =====")

    //fmt.Println(strings.Join(orgInfo.types.sourceLines(""), "\n"))
    //fmt.Println(strings.Join(newInfo.types.sourceLines(""), "\n"))

	mat := MakeDiffMatrix(len(orgInfo.types.Parts), len(newInfo.types.Parts),
		func(iA, iB int) int {
			return orgInfo.types.Parts[iA].calcDiff(newInfo.types.Parts[iB])
		})
    //fmt.Println(mat.PrettyString())
	_, rows, cols := GreedyMatch(mat, func(iA int) int {
        return orgInfo.types.Parts[iA].Weight()/2
    }, func(iB int) int {
        return newInfo.types.Parts[iB].Weight()/2
    })

	//	fmt.Println(rows, cols)
	for i := range rows {
		j := rows[i]
		if j < 0 {
			ShowDelWholeLine("type " + orgInfo.types.Parts[i].(*Fragment).Parts[0].sourceLines("")[0] + " ...")
		} else {
			if mat[i][j] > 0 {
                orgInfo.types.Parts[i].showDiff(newInfo.types.Parts[j])
			} //  if
		} // else
	} // for i

	for i, col := range cols {
		if col < 0 {
			ShowInsWholeLine("type " + newInfo.types.Parts[i].(*Fragment).Parts[0].sourceLines("")[0] + " ...")
		} // if
	} // for i
}

func DiffVars(orgInfo, newInfo *FileInfo) {
	fmt.Println("===== VARS")
	defer fmt.Println("      VARS =====")

    //fmt.Println(strings.Join(orgInfo.vars.sourceLines(""), "\n"))
    //fmt.Println(strings.Join(newInfo.vars.sourceLines(""), "\n"))

	mat := MakeDiffMatrix(len(orgInfo.vars.Parts), len(newInfo.vars.Parts), func(iA, iB int) int {
		return orgInfo.vars.Parts[iA].calcDiff(newInfo.vars.Parts[iB])
	})
    //fmt.Println(mat.PrettyString())

	_, matA, matB := GreedyMatch(mat, func(iA int) int {
        return orgInfo.vars.Parts[iA].Weight()/2
    }, func(iB int) int {
        return newInfo.vars.Parts[iB].Weight()/2
    })
    //fmt.Println(orgInfo.vars.Parts[1].Weight()/2, newInfo.vars.Parts[4].Weight()/2)

	//fmt.Println(matA, matB)
	// ast.Print(newInfo.fs, newInfo.consts)
	for i := range matA {
		j := matA[i]
		if j < 0 {
			ShowDelLines(orgInfo.vars.Parts[i].sourceLines(""), 2)
			// fmt.Println()
		} else {
			if mat[i][j] > 0 {
                orgInfo.vars.Parts[i].showDiff(newInfo.vars.Parts[j])
                fmt.Println()
			} //  if
		} // else
	} // for i

	for j, col := range matB {
		if col < 0 {
			ShowInsLines(newInfo.vars.Parts[j].sourceLines(""), 2)
		} // if
	} // for i
}

func DiffFuncs(orgInfo, newInfo *FileInfo) {
	fmt.Println("===== FUNCS")
	defer fmt.Println("      FUNCS =====")
	mat := MakeDiffMatrix(len(orgInfo.funcs.Parts), len(newInfo.funcs.Parts),
        func(iA, iB int) int {
	    	return orgInfo.funcs.Parts[iA].calcDiff(newInfo.funcs.Parts[iB])
	    })
    //fmt.Println(mat.PrettyString())

	_, matA, matB := GreedyMatch(mat, func(iA int) int {
        return orgInfo.funcs.Parts[iA].Weight()/2
    }, func(iB int) int {
        return newInfo.funcs.Parts[iB].Weight()/2
    })

	//	fmt.Println(matA, matB)
	// ast.Print(newInfo.fs, newInfo.consts)
	for i := range matA {
		j := matA[i]
		if j < 0 {
			ShowDelWholeLine(orgInfo.funcs.Parts[i].oneLine())
		} else {
			if mat[i][j] > 0 {
                orgInfo.funcs.Parts[i].showDiff(newInfo.funcs.Parts[j])
			} //  if
		} // else
	} // for i

	for j, col := range matB {
		if col < 0 {
			ShowInsWholeLine(newInfo.funcs.Parts[j].oneLine())
		} // if
	} // for i
}

func Diff(orgInfo, newInfo *FileInfo) {
    DiffPackages(orgInfo, newInfo)
    DiffImports(orgInfo, newInfo)
    DiffTypes(orgInfo, newInfo)
    DiffVars(orgInfo, newInfo)
    DiffFuncs(orgInfo, newInfo)
}

func main() {
	orgFn := "godiff-new.gogo"
	newFn := "godiff.go"
//    orgFn := `F:\job\go\src\ts\timsort.go`
//    newFn := `F:\job\go\src\ts\timsortint.go`
    
    if len(os.Args) > 1 {
        orgFn = os.Args[1]
    } // if
    
    if len(os.Args) > 2 {
        newFn = os.Args[2]
    } // if

	fmt.Printf("Analyzing difference between %s and %s ...\n", orgFn, newFn)

	orgInfo, err := Parse(orgFn)
	if err != nil {
		fmt.Println(err)
		return
	} // if

	newInfo, err := Parse(newFn)
	if err != nil {
		fmt.Println(err)
		return
	} // if
    ;
    ;

	//    fmt.Println(orgInfo)
	//    fmt.Println(newInfo)

	Diff(orgInfo, newInfo)
}