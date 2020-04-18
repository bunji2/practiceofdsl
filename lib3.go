package main

import (
	"fmt"

	"github.com/mitchellh/go-z3"
)

// IntArrayVar は与えられた名前群の整数型制約変数のリストを作成する関数
func IntArrayVar(name string, num int) (r []*z3.AST) {
	for i := 0; i < num; i++ {
		r = append(r, IntVar(fmt.Sprintf("%s[%d]", name, i)))
	}
	return
}

// BoolArrayVar は与えられた名前群のブール型制約変数のリストを作成する関数
func BoolArrayVar(name string, num int) (r []*z3.AST) {
	for i := 0; i < num; i++ {
		r = append(r, BoolVar(fmt.Sprintf("%s[%d]", name, i)))
	}
	return
}

// NumArrayVar は与えられた名前群の数値型制約変数のリストを作成する関数
func NumArrayVar(name string, num int) (r []*z3.AST) {
	for i := 0; i < num; i++ {
		r = append(r, NumVar(fmt.Sprintf("%s[%d]", name, i)))
	}
	return
}

// ArrayStrings は配列の文字列表現を作成する関数
func ArrayStrings(name string, num int) (r []string) {
	for i := 0; i < num; i++ {
		r = append(r, fmt.Sprintf("%s[%d]", name, i))
	}
	fmt.Println(r)
	return
}

// ArrayString は配列の文字列表現を作成する関数
func ArrayString(name string, index int) string {
	return fmt.Sprintf("%s[%d]", name, index)
}
