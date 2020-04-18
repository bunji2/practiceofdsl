// 制約条件のテキストを golang のコードに変換するプログラム
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {
	os.Exit(run())
}

func run() int {

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage %s src.txt dst.go\n", os.Args[0])
		return 1
	}

	// 入力ファイルの読み出し
	src, err := readSrc(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	// 入力の前後に文字列を追加
	src = `package main
func main() {
ccc = NewContext()
defer ccc.Close()
` + src + "}"

	// [XXX]
	// 利用者に対しては変数名 ccc が予約語で使用禁止であることを
	// 知らせる必要がある。
	// 理想は変数名をランダム化することだが、実装は面倒である。
	// また実装の変更は lib2.go にも影響することに注意。

	// Golang の構文としてパース
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 3
	}

	//ast.Print(fset, f)

	// main 関数のステートメントリストの取得
	stmts := pickupMainStmts(f)

	// [MEMO] 今回は main 関数の中の Assert / Solve のみを
	// 変換対象とし、main 以外の他の関数は対象外とした。
	// この仕様を変更する場合は上の行を含めて全体の見直しが必要となる。

	convStmts(stmts)

	/*
		// 各ステートメントの処理
		for i, stmt := range stmts {
			switch stmt.(type) {
			case *ast.DeclStmt: // 宣言のステートメント
				ds := stmt.(*ast.DeclStmt)

				// 変数の定義か確認
				names, typ, ok := isVarDecl(ds.Decl)
				if !ok {
					// 変数の定義でない場合はなにもしない
					break
				}
				// ステートメントを書き換え
				stmts[i] = makeASTVarDecl(names, typ)

			case *ast.ExprStmt: // 式のステートメント
				es := stmt.(*ast.ExprStmt)

				if isAssert(es.X) {
					// Assert 関数のとき
					ce := es.X.(*ast.CallExpr)
					// 第一引数を変換
					ce.Args[0] = convExpr(ce.Args[0])

				} else if isSolve(es.X) {
					// Solve 関数のとき
					ce := es.X.(*ast.CallExpr)
					var args []ast.Expr
					// Solve 関数の引数で指定された Ident を文字列に変換
					for _, arg := range ce.Args {
						ident := arg.(*ast.Ident)
						args = append(args, &ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"" + ident.Name + "\"",
						})
					}
					// Solve 関数の引数を書き換え
					ce.Args = args
				}
			default:
				// ignore
			}
		}
	*/

	// ASTをファイルに保存
	err = saveSrc(os.Args[2], f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 3
	}

	return 0
}

func convStmts(stmts []ast.Stmt) {
	// 各ステートメントの処理
	for i, stmt := range stmts {
		switch stmt.(type) {
		case *ast.DeclStmt: // 宣言のステートメント
			ds := stmt.(*ast.DeclStmt)

			// 変数の定義か確認
			names, typ, ok := isVarDecl(ds.Decl)
			if !ok {
				// 変数の定義でない場合はなにもしない
				break
			}
			// ステートメントを書き換え
			stmts[i] = makeASTVarDecl(names, typ)

		case *ast.ExprStmt: // 式のステートメント
			es := stmt.(*ast.ExprStmt)

			if isAssert(es.X) {
				// Assert 関数のとき
				ce := es.X.(*ast.CallExpr)
				// 第一引数を変換
				ce.Args[0] = convExpr(ce.Args[0])

			} else if isSolve(es.X) {
				// Solve 関数のとき
				ce := es.X.(*ast.CallExpr)
				var args []ast.Expr
				// Solve 関数の引数で指定された Ident を文字列に変換
				for _, arg := range ce.Args {
					ident := arg.(*ast.Ident)
					args = append(args, &ast.BasicLit{
						Kind:  token.STRING,
						Value: "\"" + ident.Name + "\"",
					})
				}
				// Solve 関数の引数を書き換え
				ce.Args = args
			}
		case *ast.ForStmt:
			fs := stmt.(*ast.ForStmt)
			if fs.Body != nil {
				convStmts(fs.Body.List)
			}
		case *ast.RangeStmt:
			rs := stmt.(*ast.RangeStmt)
			if rs.Body != nil {
				convStmts(rs.Body.List)
			}
		case *ast.IfStmt:
			is := stmt.(*ast.IfStmt)
			if is.Body != nil {
				convStmts([]ast.Stmt{is.Body})
			}
			if is.Else != nil {
				convStmts([]ast.Stmt{is.Else})
			}
		case *ast.BlockStmt:
			bs := stmt.(*ast.BlockStmt)
			convStmts(bs.List)
		default:
			// ignore
		}
	}

}

// isVarDecl は与えられた AST が変数宣言かどうかをチェックする関数。
// 返り値は宣言されている変数の名前のリストと、型を示す文字列。
func isVarDecl(decl ast.Decl) (names []string, typ string, ok bool) {
	var gd *ast.GenDecl

	// GenDecl か？
	gd, ok = decl.(*ast.GenDecl)
	if !ok {
		return
	}
	ok = false

	// token.VAR か？
	if gd.Tok != token.VAR {
		return
	}

	var vs *ast.ValueSpec

	// ValueSpec か？
	vs, ok = gd.Specs[0].(*ast.ValueSpec)
	if !ok {
		return
	}
	ok = false

	// vs.Type は Ident もしくは ArrayType か？
	switch vs.Type.(type) {
	case *ast.Ident:
		tmp := vs.Type.(*ast.Ident)

		// Int, Num, Bool のいずれか？
		switch tmp.Name {
		case "Int":
		case "Num":
		case "Bool":
		default:
			// 上記以外
			return
		}
		typ = tmp.Name

	case *ast.ArrayType:
		tmp := vs.Type.(*ast.ArrayType)
		var elt *ast.Ident
		elt, ok = tmp.Elt.(*ast.Ident)
		if !ok {
			// struct などは対象外
			return
		}
		ok = false

		// Int, Num, Bool のいずれか？
		switch elt.Name {
		case "Int":
		case "Num":
		case "Bool":
		default:
			// 上記以外
			return
		}

		var len *ast.BasicLit
		len, ok = tmp.Len.(*ast.BasicLit)
		if !ok {
			return
		}
		ok = false

		// [XXX] ↑ BasicLit 固定に注意。つまり、配列宣言の要素数は固定。
		// var xs [5]Int // これは OK
		//
		// 次のような要素数を変数にするような場合はランタイムエラーとなる。
		// n := 5
		// var xs [n]Int // これは NG

		typ = elt.Name + "_" + len.Value
		// {Int,Num,Bool}_要素数

	default:
		// 上記以外
		return
	}

	// 定義されている名前をリスト化
	for _, ident := range vs.Names {
		names = append(names, ident.Name)
	}

	ok = true
	return
}

// makeASTVarDecl は制約変数を定義するASTを生成する関数
func makeASTVarDecl(names []string, typ string) ast.Stmt {
	// Before: var x, y Int
	// After:  x, y := IntVar("x"), IntVar("y")

	typs := strings.Split(typ, "_")
	if len(typs) > 1 {
		// "typ_num" の場合は、配列型となる
		num, _ := strconv.Atoi(typs[1])
		return makeASTVarArrayDecl(names, typs[0], num)
	}

	n0 := ast.NewIdent(typ + "Var")
	var n1, n3 []ast.Expr
	for _, name := range names {
		n1 = append(n1, ast.NewIdent(name))
		n2 := &ast.BasicLit{Value: "\"" + name + "\"", Kind: token.STRING}
		n3 = append(n3, &ast.CallExpr{Fun: n0, Args: []ast.Expr{n2}})
	}
	n4 := &ast.AssignStmt{Lhs: n1, Tok: token.DEFINE, Rhs: n3}
	return n4
}

// makeASTVarArrayDecl は配列の制約変数を定義するASTを生成する関数
func makeASTVarArrayDecl(names []string, typ string, num int) ast.Stmt {
	n0 := ast.NewIdent(typ + "ArrayVar")
	n3 := &ast.BasicLit{Value: fmt.Sprintf("%d", num), Kind: token.INT}
	var n1, n4 []ast.Expr
	for _, name := range names {
		n1 = append(n1, ast.NewIdent(name))
		n2 := &ast.BasicLit{Value: fmt.Sprintf("\"%s\"", name), Kind: token.STRING}
		n4 = append(n4, &ast.CallExpr{Fun: n0, Args: []ast.Expr{n2, n3}})
	}
	n5 := &ast.AssignStmt{Lhs: n1, Tok: token.DEFINE, Rhs: n4}
	return n5
}

// convExpr は Assert 関数の引数で指定された式のASTを変換する関数
func convExpr(expr ast.Expr) (r ast.Expr) {
	//fmt.Println("convExpr: expr=", expr)
	switch expr.(type) {
	case *ast.BinaryExpr:
		r = convBinaryExpr(expr.(*ast.BinaryExpr))
	case *ast.UnaryExpr:
		r = convUnaryExpr(expr.(*ast.UnaryExpr))
	case *ast.CallExpr:
		r = convCallExpr(expr.(*ast.CallExpr))
	case *ast.ParenExpr:
		r = convExpr(expr.(*ast.ParenExpr).X)
	case *ast.Ident:
		r = convIdent(expr.(*ast.Ident))
	case *ast.BasicLit:
		//fmt.Println("convExpr: basiclit expr=", expr)
		r = convBasicLit(expr.(*ast.BasicLit))
	default:
		// 上記以外は変換しない
		r = expr
		//fmt.Println("convExpr: other expr=", expr)
	}
	return
}

// convBinaryExpr は二項演算式を変換する関数
func convBinaryExpr(expr *ast.BinaryExpr) (r ast.Expr) {
	//fmt.Println("convBinaryExpr: expr=", expr)

	// Before: expr1 + expr2
	// After:  conv(expr1).Add(conv(expr2))

	// Before: expr1 == expr2
	// After:  conv(expr1).Eq(conv(expr2))

	// Before: expr1 != expr2
	// After:  (conv(expr1).Eq(conv(expr2))).Not()

	var op string
	switch expr.Op {
	case token.ADD: // +
		op = "Add"
	case token.SUB: // -
		op = "Sub"
	case token.MUL: // *
		op = "Mul"
	case token.REM: // %
		op = "Mod"
	case token.LAND: // &&
		op = "And"
	case token.LOR: // ||
		op = "Or"
	case token.XOR: // ^
		op = "Xor"
	case token.GTR: // >
		op = "Gt"
	case token.GEQ: // >=
		op = "Ge"
	case token.LSS: // <
		op = "Lt"
	case token.LEQ: // <=
		op = "Le"
	case token.EQL: // ==
		op = "Eq"
	case token.NEQ: // != // [NEQ]
		op = "Eq"
	default:
		r = expr
		return
	}

	//fmt.Println("op =", op)

	x := convExpr(expr.X)
	y := convExpr(expr.Y)

	r = &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   x,
			Sel: ast.NewIdent(op),
		},
		Args: []ast.Expr{
			y,
		},
	}
	if expr.Op == token.NEQ { // [NEQ]
		r = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   r,
				Sel: ast.NewIdent("Not"),
			},
		}
	}
	return
}

// convUnaryExpr は単行演算式を変換する関数
func convUnaryExpr(expr *ast.UnaryExpr) (r ast.Expr) {
	// Before: !expr
	// After:  conv(expr).Not()

	// Before: -expr
	// After:  conv(expr).Neg()

	var ident *ast.Ident
	switch expr.Op {
	case token.NOT: // !
		ident = ast.NewIdent("Not")
	case token.SUB: // -
		ident = ast.NewIdent("Neg")
	default:
		// 上記以外は変換せずリターン
		r = expr
		return
	}

	r = &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   convExpr(expr.X),
			Sel: ident,
		},
	}

	return
}

// convCallExpr は関数呼び出し式を変換する関数
func convCallExpr(expr *ast.CallExpr) (r ast.Expr) {
	// Before: Distinct(x1,x2,...,xN)
	// After:  x1.Distinct(x2,...,xN)

	// Before: expr1.Implies(expr2)
	// After:  conv(expr1).Implies(conv(expr2))

	// 引数の AST を変換
	var args []ast.Expr
	for _, arg := range expr.Args {
		args = append(args, convExpr(arg))
	}

	// 関数呼び出し式のケースで分岐
	switch expr.Fun.(type) {
	case *ast.Ident: // x (args)
		ident := expr.Fun.(*ast.Ident)
		switch ident.Name {
		case "Distinct":
			r = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   args[0],
					Sel: ast.NewIdent("Distinct"),
				},
				Args: args[1:],
			}
			// [TODO] Distinct の引数は変数のみなので、
			// args の各要素がすべて変数の AST かどうかチェックすべき。
			// 変数以外の式が引数に指定された場合は多分、
			// ランタイムエラーになると思われる。
		default:
			// Distinct 関数以外は変換しない
			r = expr
		}

	case *ast.SelectorExpr: // x.y (args)
		se := expr.Fun.(*ast.SelectorExpr)
		switch se.Sel.Name {
		case "Implies":
		case "Iff":
		case "Ite":
		case "Pow":
		default:
			// 上記以外は変換せずリターン
			r = expr
			return
		}
		r = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   convExpr(se.X),
				Sel: se.Sel, // Implies, Iff, Ite, or Pow
			},
			Args: args,
		}

	default:
		// 上記以外のケースは変換せずリターン
		r = expr
	}

	return
}

// convBasicLit は基本リテラルを変換する関数
func convBasicLit(expr *ast.BasicLit) (r ast.Expr) {
	//fmt.Println("convBasicLit: expr=", expr)
	// Before: 123
	// After:  IntVal(123)

	// Before: 123.456
	// After:  NumVal("123.456")

	switch expr.Kind {
	case token.INT:
		r = &ast.CallExpr{
			Fun: ast.NewIdent("IntVal"),
			Args: []ast.Expr{
				expr,
			},
		}
	case token.FLOAT:
		r = &ast.CallExpr{
			Fun: ast.NewIdent("NumVal"),
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + expr.Value + "\"",
				},
			},
		}
	default:
		// 上記以外は変換しない
		r = expr
	}

	return
}

// convIdent は識別子を変換する関数。識別子のうち真理値（true or false）が該当。
func convIdent(expr *ast.Ident) (r ast.Expr) {
	// Before: true
	// After:  True()

	// Before: false
	// After:  False()

	var name string
	switch expr.Name {
	case "true":
		name = "True"
	case "false":
		name = "False"
	default:
		// 変換せずリターン
		r = expr
		return
		// [XXX]
		// Assert 関数の式の中に現れる識別子が制約変数ではない変数が
		// 現れるケースを想定していないが問題があるかもしれない。そうでないかもしれない。
		// 制約変数の宣言の変換を行う際に制約変数名のリストを作成し、
		// それに含まれるかどうかを convIdent でチェックすることも考えてみるべきかもしれない。
	}

	r = &ast.CallExpr{
		Fun: ast.NewIdent(name),
	}
	return
}

// isAssert は式が Assert 関数かどうかをチェックする関数
func isAssert(expr ast.Expr) bool {
	//fmt.Println("# isAssert")
	ce, ok := expr.(*ast.CallExpr)
	if ok {
		// identifier (args) の形の関数呼び出しか
		ident, ok := ce.Fun.(*ast.Ident)
		if ok && ident.Name == "Assert" {
			// 引数の数は一つか？
			if len(ce.Args) == 1 {
				return true
			}
		}
	}
	return false
}

// isSolve は式が Solve 関数かどうかをチェックする関数
func isSolve(expr ast.Expr) bool {
	//fmt.Println("# isSolve")
	ce, ok := expr.(*ast.CallExpr)
	if ok {
		// identifier (args) の形の関数呼び出しか
		ident, ok := ce.Fun.(*ast.Ident)
		if ok && ident.Name == "Solve" {
			// 引数はすべて Ident か？
			for _, arg := range ce.Args {
				_, ok := arg.(*ast.Ident)
				if !ok {
					return false
				}
			}
			return true
		}
	}
	return false
}

// pickupMainStmts は main 関数のステートメントリストを取得する関数
func pickupMainStmts(fileNode *ast.File) (stmts []ast.Stmt) {
	// ファイルノードのトップレベルの「宣言」の中から main 関数を
	// 見つけ出し、そのステートメントリストを抽出する。
	for _, n := range fileNode.Decls {
		// 関数宣言のうちその名前が "main" のものをみつける
		funcDecl, ok := n.(*ast.FuncDecl)
		if ok && funcDecl.Name.Name == "main" {
			stmts = funcDecl.Body.List
			break
		}
	}
	return
}

// saveSrc は AST をファイルに保存する関数
func saveSrc(filename string, f *ast.File) (err error) {
	var w *os.File
	w, err = os.Create(filename)
	if err != nil {
		return
	}
	defer w.Close()
	// AST をファイルに保存
	format.Node(w, token.NewFileSet(), f)
	return
}

// readSrc はファイルを読み出す関数
func readSrc(filename string) (string, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
