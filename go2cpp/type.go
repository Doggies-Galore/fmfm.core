package go2cpp

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func (g *generator) isStructType(expr ast.Expr) bool {
	return g.isStructTypeImpl(g.info.TypeOf(expr))
}

func (g *generator) isStructTypeImpl(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Struct:
		return true
	case *types.Named:
		return g.isStructTypeImpl(t.Underlying())
	}
	return false
}

func (g *generator) isArrayType(expr ast.Expr) bool {
	return g.isArrayTypeImpl(g.info.TypeOf(expr))
}

func (g *generator) isArrayTypeImpl(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Array:
		return true
	case *types.Named:
		return g.isArrayTypeImpl(t.Underlying())
	}
	return false
}

func (g *generator) isSliceType(expr ast.Expr) bool {
	return g.isSliceTypeImpl(g.info.TypeOf(expr))
}

func (g *generator) isSliceTypeImpl(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Slice:
		return true
	case *types.Named:
		return g.isSliceTypeImpl(t.Underlying())
	}
	return false
}

func (g *generator) isMapType(expr ast.Expr) bool {
	return g.isMapTypeImpl(g.info.TypeOf(expr))
}

func (g *generator) isMapTypeImpl(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Map:
		return true
	case *types.Named:
		return g.isMapTypeImpl(t.Underlying())
	}
	return false
}

func (g *generator) isStringType(typ types.Type) bool {
	return typ != nil && typ.String() == "string"
}

func (g *generator) dumpZeroValue(writer io.Writer, typ types.Type) {
	switch typ.String() {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(writer, "%s(0)", typ.String())
	case "float32", "float64":
		fmt.Fprintf(writer, "%s(.0)", typ.String())
	default:
		fmt.Fprintf(writer, "{} /*%s*/", typ.String())
	}
}

var nonAlphaNumRe = regexp.MustCompile(`\W`)

func (g *generator) typeIdent(typ types.Type) string {
	n, s, ok := g.formatTypeImpl(typ, true)
	if !ok {
		return "UNKNOWN"
	}
	s = fmt.Sprintf(s, "")
	return n + nonAlphaNumRe.ReplaceAllStringFunc(s, func(s string) string {
		switch s {
		case "*":
			return "Ptr"
		case "[":
			return "Arr"
		case "]":
			return ""
		default:
			return s
		}
	})
}

func (g *generator) packageAndType(t *types.Named) (string, string) {
	segs := strings.Split(t.String(), "/")
	i := len(segs) - 1
	p := strings.Split(segs[i], ".")
	segs[i] = strings.Join(p[:len(p)-1], ".")
	return strings.Join(segs, "/"), p[len(p)-1]
}

func (g *generator) formatType(typ types.Type) (string, string, bool) {
	return g.formatTypeImpl(typ, false)
}

func (g *generator) formatTypeImpl(typ types.Type, barePtr bool) (string, string, bool) {
	switch t := typ.(type) {
	case nil:
		return "auto", "%s", true
	case *types.Basic:
		return t.Name(), "%s", true
	case *types.Named:
		// @todo インポート名がローカル変数でマスクされたときの対処
		pkg, typeName := g.packageAndType(t)
		return g.localPkgPrefix(pkg) + typeName, "%s", true
	case *types.Pointer:
		n, s, ok := g.formatType(t.Elem())
		if barePtr {
			return n, "*" + s, ok
		}
		s = fmt.Sprintf(s, "")
		return fmt.Sprintf("std::shared_ptr<%s%s>", n, s), "%s", ok
	case *types.Array:
		n, s, ok := g.formatType(t.Elem())
		return n, s + "[" + strconv.Itoa(int(t.Len())) + "]", ok
	case *types.Slice:
		n, s, ok := g.formatType(t.Elem())
		s = fmt.Sprintf(s, "")
		return fmt.Sprintf("std::vector<%s%s>", n, s), "%s", ok
	default:
		return g.debugSInspect(typ, "formatType"), "", false
	}
}
