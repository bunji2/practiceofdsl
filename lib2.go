package main

import (
	"github.com/mitchellh/go-z3"
)

// グローバル変数
var ccc Context

// IntVar は整数型の制約変数を作成する関数
func IntVar(name string) *z3.AST {
	return ccc.IntVar(name)
}

// IntVal は整数値を作成する関数
func IntVal(value int) *z3.AST {
	return ccc.IntVal(value)
}

// BoolVar はブール型の制約変数を作成する関数
func BoolVar(name string) *z3.AST {
	return ccc.BoolVar(name)
}

// NumVar は数値型の制約変数を作成する関数
func NumVar(name string) *z3.AST {
	return ccc.NumVar(name)
}

// NumVal は数値のASTノードを作成する関数
func NumVal(value string) *z3.AST {
	return ccc.NumVal(value)
}

/*
// FloatVar は浮動小数点型の制約変数を作成する関数
func FloatVar(name string) *z3.AST {
	return ccc.FloatVar(name)
}

// FloatVal は整数値を作成する関数
func FloatVal(value float64) *z3.AST {
	return ccc.FloatVal(value)
}
*/

// Assert は制約条件を宣言する関数
func Assert(cond *z3.AST) {
	ccc.Assert(cond)
}

// Solve は制約を解決する変数の値を表示する関数
func Solve(names ...string) {
	ccc.Solve(names...)
}

// True は True 値のASTノードを作成する関数
func True() *z3.AST {
	return ccc.True()
}

// False は False 値のASTノードを作成する関数
func False() *z3.AST {
	return ccc.False()
}
