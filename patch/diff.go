package z3

// #include <stdlib.h>
// #include "go-z3.h"
import "C"
import "unsafe"

// Neg creates an AST node representing -(a)
//
// Maps to: Z3_mk_unary_minus
func (a *AST) Neg() *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_unary_minus(a.rawCtx, a.rawAST),
	}
}

// Pow creates an AST node representing adding.
//
// Maps to: Z3_mk_power
func (a *AST) Pow(arg *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_power(
			a.rawCtx,
			a.rawAST,
			arg.rawAST,
		),
	}
}

// Mod creates an AST node representing arg1 mod arg2.
//
// Maps to: Z3_mk_mod
func (a *AST) Mod(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_mod(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

//

// RealSort returns the int type.
func (c *Context) RealSort() *Sort {
	return &Sort{
		rawCtx:  c.raw,
		rawSort: C.Z3_mk_real_sort(c.raw),
	}
}

// Num create a numeral of a given sort.
//
// Maps: Z3_mk_numeral
func (c *Context) Num(v string, typ *Sort) *AST {
	cv := C.CString(v)
	defer C.free(unsafe.Pointer(cv))
	aa := C.Z3_mk_numeral(c.raw, cv, typ.rawSort)
	return &AST{
		rawCtx: c.raw,
		rawAST: aa,
	}
}
