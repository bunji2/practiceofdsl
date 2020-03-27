# Practice of DSL

ここでは、DSL のデザインに関するメモを記していく。

----

## はじめに

DSL とは Domain Specific Language の略であり日本語では「ドメイン固有言語」と言うらしい。

汎用的なプログラミング言語は DSL とは呼ばない。あくまで用途が限定されるような場合に DSL と呼ぶようだ。

ここでは、SMT Solver を使うための DSL を考える。

## 基本方針

SMT Solver 向けの DSL を考えるにあたって次の方針とする。

* 実装に使用するプログラミング言語は golang とする。
* 新しく作る DSL の構文仕様は golang の構文仕様と親和性のあるものとする。
* SMT Solver として "Z3" を使うが、フロントエンドは go-binding である "go-z3" を使う。
* "go-z3" 以外は、すべて標準パッケージのみを使って実装する。
* 新たにスクラッチからパーサを実装することはしない。

今回の DSL の利用者は以下を想定する。

* 論理学の知識があり SMT Solver の制約条件の検討ができる
* 基本的な IT リテラシはあるものとし、OS などのプラットフォームの基本操作は一通りできる。
* プログラミングスキルはない。
* 手順書や業務手順書が与えられれば、未知のシステムでも手順通りに操作する勤勉さはある。

制約式の検討はできてもプログラミングスキルがない。このような条件下で運用していくには、DSL が適切な抽象レベルで適切なパワーを持つような工夫をとることがキモとなる。

----

## ステップ0. 最初のサンプルコード

SMT Solver とは Satisfiable Modulo Theories Solver の略で、一階述語論理式で記述された制約条件を満たす値を解決してくれるシステムである。

例題として、次のような条件式を満たす整数 x と y の解決を例とする。

![](https://latex.codecogs.com/gif.latex?x&plus;y=24\wedge{x-y=2})

この例を解決するサンプルコードを示す。

```golang
package main

import (
	"fmt"
	"github.com/mitchellh/go-z3"
)

func main() {
	// コンテクストの作成
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	defer ctx.Close()

	// ソルバーの作成
	solver := ctx.NewSolver()
	defer solver.Close()

	// 制約変数の定義
	x := ctx.Const(ctx.Symbol("x"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("y"), ctx.IntSort())

	// 制約条件
	// x + y = 24
	solver.Assert(x.Add(y).Eq(ctx.Int(24, ctx.IntSort())))
	// x - y = 2
	solver.Assert(x.Sub(y).Eq(ctx.Int(2, ctx.IntSort())))

	// 解決可能化チェック
	if v := solver.Check(); v != z3.True {
		fmt.Println("解決不能")
		return
	}

	// 結果の表示
	m := solver.Model()
	values := m.Assignments()
	m.Close()
	fmt.Printf("x = %s\n", values["x"])
	fmt.Printf("y = %s\n", values["y"])
}
```

実行結果：

```
% go run sample.go 
x = 13
y = 11
```

実行すると制約条件を満たす整数 x と y の値が表示される。

実際はこの例題のような一次方程式を解くような単純なものだけでなく、もっと複雑な制約条件を扱うわけだが今回は省略する。

さて、利用者の立場で考えてみると、上のサンプルコードのうち SMT Solver に渡す制約条件の記述については興味はあるが、専門外のライブラリのインポート・コンテクストやソルバーの作成・制約変数の定義など、SMT Solver を動かすためのコードの記述は意味不明であり、煩わしいだけだ。

![fig01.png](fig/fig01.png)

以降は上のサンプルコードをステップ0 として、ステップ・バイ・ステップで DSL 化を勧めていくことにする。

まずは、上のコードの雑多な部分をライブラリ化することで、どれだけ記述が簡単になるかをみてみる。


----

## ステップ1. ライブラリ化

ステップ1 のサンプルコード sample1.go を示す。

```
package main

func main() {
	// コンテクストとソルバーの作成
	c := NewContext()
	defer c.Close()

	// 制約変数
	x := c.IntVar("x")
	y := c.IntVar("y")

	// 制約条件
	// x + y = 24
	c.Assert(x.Add(y).Eq(c.IntVal(24)))
	// x - y = 2
	c.Assert(x.Sub(y).Eq(c.IntVal(2)))

	// 解決結果の表示
	c.Solve("x", "y")
}
```

前ステップのコードからの変更を示す。

![fig02.png](fig/fig02.png)

次のように実行する。

```
% go run sample1.go lib.go
x = 13
y = 11
```

作成したライブラリ [lib.go](lib.go) のポイントは下の通りである。

* z3 のコンテクストやソルバーをメンバーとしてもつ構造体型 Context 型の導入する。
* Context 型のメソッドとして、変数定義などのサンプルコードで使用する関数を保持する。
* go-z3 への依存性をすべて lib.go に寄せることにより、サンプルコードからの go-z3 のインポートを不要にする。


ステップ1 はステップ0 よりも、ライブラリ化によって利用者が記述するコード量は減少する。
しかしそれでも golang 特有のパッケージ宣言や main 関数の宣言など、毎回同じ内容を記述するのは無駄が多いのでなくしてしまいたい。

![fig04](fig/fig04.png)

----

## ステップ2. 差分テキスト化

ステップ2 のサンプルコード sample2.txt を示す。

```
// 制約変数
x := c.IntVar("x")
y := c.IntVar("y")

// 制約条件
// x + y = 24
c.Assert(x.Add(y).Eq(c.IntVal(24)))
// x - y = 2
c.Assert(x.Sub(y).Eq(c.IntVal(2)))

// 解決結果の表示
c.Solve("x", "y")
```

前ステップのコードからの変更を示す。

![fig05](fig/fig05.png)

次のように実行する。

```
% run.sh sample2.txt
x = 13
y = 11
```

使用するシェルスクリプト run.sh 内の処理は次の通り。

![fig06](fig/fig06.png)

run.sh の中で golang 特有のパッケージ宣言や main 関数の宣言を補完し、go run コマンドで実行する。

具体的な実装は次の通りである。

```
#!/bin/sh

filename=`basename $1 .txt`$$.go

(
    echo "package main"
    echo "func main() {"
    echo "c := NewContext()"
    echo "defer c.Close()"
    cat $1
    echo "}"
) > $filename

go run $filename lib.go

rm $filename
```

ステップ2 では差分テキスト化することによって golang 特有のパッケージ宣言や main 関数の宣言などがなくなり、利用者の記述量はさらに減少した。

しかし差分テキスト化の副作用として、各関数のプレフィクスの "c." がもはや意味をなさなくなってしまった。
無駄なのでこれをなくしてしまいたい。

![fig07](fig/fig07.png)

----

## ステップ3. ライブラリ化その２

ステップ3 のサンプルコード sample3.txt を示す。

```
// 制約変数
x := IntVar("x")
y := IntVar("y")

// 制約条件
// x + y = 24
Assert(x.Add(y).Eq(IntVal(24)))
// x - y = 2
Assert(x.Sub(y).Eq(IntVal(2)))

// 解決結果の表示
Solve("x", "y")
```

前ステップのコードからの変更を示す。

![fig08](fig/fig08.png)

実行はステップ2 と同じである。

```
% run.sh sample3.txt
x = 13
y = 11
```

run.sh 內部の処理は次の通り。

![fig09](fig/fig09.png)

ステップ2 との run.sh の差分は、ライブラリ [lib2.go](lib2.go) が増えたことである。
追加したライブラリにより、コンテクスト変数のグローバル化を行なった。

```
#!/bin/sh

filename=`basename $1 .txt`$$.go

(
    echo "package main"
    echo "func main() {"
    echo "ccc = NewContext()"
    echo "defer ccc.Close()"
    cat $1
    echo "}"
) > $filename

go run $filename lib.go lib2.go

rm $filename
```

無駄な "c." プレフィクスがなくなって、利用者の記述量はまた低減された。

しかしそれでも、制約条件が直感的ではないという問題が残っている。

利用者としては「数学的な条件式」を使いたい。

この他、制約変数で使っている文字列表記は冗長なので単純化したい。

![fig10](fig/fig10.png)

----

## ステップ4. 制約条件の数式化

ステップ4 のサンプルコード sample4.txt を示す。

```golang
// 制約変数
var x, y Int

// 制約条件
Assert(x + y == 24)
Assert(x - y == 2)

// 解決結果の表示
Solve(x, y)
```

前ステップのコードからの変更を示す。

![fig11](fig/fig11.png)

上のような変更を行なうには Assert 関数の引数の式の構造を自動的に変換する必要があるが、
ライブラリ化や差分テキスト化だけでは対応できない。

今回は Assert 関数の引数の式の "AST" を加工することで対応した。

"AST" とは "Abstract Syntax Tree" の略であり、日本語では「抽象構文木」と呼ばれる。
やや端折って簡単に言い切ってしまうと、ソースコードの構文に対応する木構造のことである。

シェルスクリプト run.sh 內部の処理を示す。

![fig12](fig/fig12.png)


```
#!/bin/sh

filename=`basename $1 .txt`.go

conv $1 $filename

go run $filename lib.go lib2.go

rm $filename
```

conv コマンドの中で変換している AST の前後を以下に示す。

![fig1](fig1.png)

(図中の同じ色の箇所は前後で対応する箇所)


以下、conv コマンドにおけるコードを抜粋して処理を解説する。

### 差分テキストの補完処理とパージング処理

```golang
// 入力コードの読み出し
src := readSrc(os.Args[1])

// 差分テキストの補完処理。入力コードの前後に補完コードをコンカテネーション
src = `package main
func main() {
ccc = NewContext()
defer ccc.Close()` + src + "}"

// Golang のソースコードとしてパースし、AST を取得
fileNode, err := parser.ParseFile(fset, "", src, 0)
```

差分テキストの補完処理は単に補完コードをコンカテネーションするだけである。
補完コードのパージング処理には go の標準パッケージである go/parser パッケージを利用する。
このようにしてパージング処理の結果、上の例では fileNode 変数に AST が格納される。


### 変換箇所の特定

```golang

// 各ステートメントの処理
for i, stmt := range stmts {
	switch stmt.(type) {
	...
	case *ast.ExprStmt: // 式のステートメント
		es := stmt.(*ast.ExprStmt)
		if isAssert(es.X) { // "Assert" 関数のとき
			ce := es.X.(*ast.CallExpr)
			// 第一引数を書き換え
			ce.Args[0] = convExpr(ce.Args[0])
```

各ステートメントの中から Assert 関数のステートメントをみつけ、変換の対象となる式を特定する。


### 式のASTの変換

```golang
// convExpr は Assert 関数の引数で指定された式のASTを変換する関数
func convExpr(expr ast.Expr) (r ast.Expr) {
	switch expr.(type) {
	case *ast.BinaryExpr:
		r = convBinaryExpr(expr.(*ast.BinaryExpr))	// 二項演算式の変換
	case *ast.UnaryExpr:
		r = convUnaryExpr(expr.(*ast.UnaryExpr))	// 単項演算式の変換
	case *ast.CallExpr:
		r = convCallExpr(expr.(*ast.CallExpr))		// 関数呼び出し式の変換
	case *ast.ParenExpr:
		r = convExpr(expr.(*ast.ParenExpr).X)		// 括弧で囲まれた式の変換
	case *ast.Ident:
		r = convIdent(expr.(*ast.Ident))		// 識別子からなる式の変換
	case *ast.BasicLit:
		r = convBasicLit(expr.(*ast.BasicLit))		// 整数などのリテラルからなる式の変換
	default:
		// 上記以外は変換しない。
		r = expr
	}
	return
}
```

「式」には、二項演算式、単項演算式、関数呼び出し、など、複数のケースがあるため、
それぞれのケースに応じて switch 文で分岐しながら AST の木構造をトラバースしていくわけだが、
変換を必要としない式もあることに注意。

```golang
// convUnaryExpr は単行演算式を変換する関数
func convUnaryExpr(expr *ast.UnaryExpr) (r ast.Expr) {
	if expr.Op != token.NOT {
		r = expr
		return
	}
	r = &ast.CallExpr{	// 対応する AST ノードの生成
		Fun: &ast.SelectorExpr{
			X:   convExpr(expr.X), // NOT演算子の引数の変換（再帰的な変換）
			Sel: ast.NewIdent("Not"),
		},
	}
	return
}
```

対応する AST ノードを生成することで式を変換していく。

Assert 関数の引数の変換の他に、次の変数宣言と Solve 関数の引数の変換も行なう。

### 変数宣言の変換

制約変数の変数宣言では通常の "int" の変数宣言と区別がつくよう、"Int" という型で表現する。
制約変数の宣言は IntVar 関数を用いた表現に変換することでコンパイルできるようにする。

```
// 変換前
var x, y Int
// 変換後
x, y := IntVar("x"), IntVar("y")|
```

変換前後の AST を示す。

![AST2](fig/AST2.png)

### Solve関数の引数の変換

表示する制約変数の指定時に引用符をつけなくてもよいようにしたので、Solve 関数の引数に並ぶ変数名を文字列に変換する。

```
// 変換前
Solve(x, y)
// 変換後
Solve("x", "y")|
```

変換前後の AST を示す。

![AST3](fig/AST3.png)


## おわりに

[](
## 評価

各ステップでのサンプルコードの行数・バイト数、および、実装したライブラリ等の行数をそれぞれまとめる。

| | code [lines] | code [bytes] | libs [lines] |
|:--|:---:|:--:|:--:|
|STEP0|41|806|-|
|STEP1|19|335|80|
|STEP2|11|212|95|
|STEP3|11|198|133|
|STEP4|8|122|525|

ステップ4のサンプルコードの行数・バイト数はそれぞれステップ0の20%・15%に低減された。

他方で各ステップを経るほど実装したライブラリ等の行数は増えることがわかる。

特にステップ4では AST の変換処理が実装量の大勢を占める。
それでも、これらは一度実装してしまえ今回のサンプルコード以外でも再利用可能な部分であり、無駄な投資とはならない。

)
