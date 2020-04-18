package main

import (
	"fmt"

	"github.com/mitchellh/go-z3"
)

// Context は z3 のコンテクストとソルバーを保持する構造体型
type Context struct {
	ctx    *z3.Context
	solver *z3.Solver
	vars   map[string]bool
}

// NewContext は新しいコンテクストを生成する関数
func NewContext() Context {
	// コンテクストの作成
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	return Context{
		ctx:    ctx,
		solver: ctx.NewSolver(),
		vars:   map[string]bool{},
	}
}

// Close はコンテクストをクローズする関数
func (c Context) Close() {
	if c.solver != nil {
		c.solver.Close()
	}
	if c.ctx != nil {
		c.ctx.Close()
	}
}

// BoolVar はブール型の制約変数のASTノードを作成する関数
func (c Context) BoolVar(name string) *z3.AST {
	c.vars[name] = true
	return c.ctx.Const(c.ctx.Symbol(name), c.ctx.BoolSort())
}

// IntVar は整数型の制約変数のASTノードを作成する関数
func (c Context) IntVar(name string) *z3.AST {
	c.vars[name] = true
	return c.ctx.Const(c.ctx.Symbol(name), c.ctx.IntSort())
}

// IntVal は整数値のASTノードを作成する関数
func (c Context) IntVal(value int) *z3.AST {
	return c.ctx.Int(value, c.ctx.IntSort())
}

// NumVar は数値型の制約変数のASTノードを作成する関数
func (c Context) NumVar(name string) *z3.AST {
	c.vars[name] = true
	return c.ctx.Const(c.ctx.Symbol(name), c.ctx.RealSort())
}

// NumVal は数値のASTノードを作成する関数
func (c Context) NumVal(value string) *z3.AST {
	return c.ctx.Num(value, c.ctx.RealSort())
}

/*
// NewVar は指定されたソートの制約変数のASTノードを作成する関数
func (c Context) NewVar(name string, idx int, sort *z3.Sort) *z3.AST {

	c.vars[name] = idx+1
	return c.ctx.Const(c.ctx.Symbol(name), sort)
}
*/

/*
// FloatVar は浮動小数点型の制約変数のASTノードを作成する関数
func (c Context) FloatVar(name string) *z3.AST {
	return c.ctx.Const(c.ctx.Symbol(name), c.ctx.FloatSort())
}

// FloatVal は浮動小数点のASTノードを作成する関数
func (c Context) FloatVal(value float64) *z3.AST {
	return c.ctx.Float(value, c.ctx.FloatSort())
}
*/

// Assert は制約条件を宣言する関数
func (c Context) Assert(cond *z3.AST) {
	c.solver.Assert(cond)
}

// Solve は制約を解決する変数の値を表示する関数
func (c Context) Solve(names ...string) {
	// 解決可能かどうかを調べる
	if v := c.solver.Check(); v != z3.True {
		fmt.Println("unsolvable")
		return
	}

	// 制約を満たす値の取得
	m := c.solver.Model()
	values := m.Assignments()
	m.Close()

	// 可変引数で指定された変数名の値を表示
	for _, name := range names {
		//fmt.Println("name =", name)
		if c.vars[name] {
			fmt.Printf("%s = %s\n", name, values[name].String())
		} else {
			// 配列の可能性
			i := 0
			for {
				idxName := fmt.Sprintf("%s[%d]", name, i)
				if c.vars[idxName] {
					fmt.Printf("%s = %s\n", idxName, values[idxName].String())
				} else {
					break
				}
				i++
			}
		}
		//fmt.Printf("%s = %s.\n", name, values[name].FString2())
	}
}

// True は True 値のASTノードを作成する関数
func (c Context) True() *z3.AST {
	return c.ctx.True()
}

// False は False 値のASTノードを作成する関数
func (c Context) False() *z3.AST {
	return c.ctx.False()
}
