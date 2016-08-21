// Copyright The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

// Package ast declares the types used to represent syntax trees for Go
// packages.

// ast 包声明了用于描述 Go packages 语法树的类型.
package ast

import (
    "bytes"
    "fmt"
    "go/scanner"
    "go/token"
    "io"
    "os"
    "reflect"
    "sort"
    "strconv"
    "strings"
    "unicode"
    "unicode/utf8"
)

// The list of possible Object kinds.

// Object kinds 可许值列表.
const (
    Bad ObjKind = iota // for error handling
    Pkg                // package
    Con                // constant
    Typ                // type
    Var                // variable
    Fun                // function or method
    Lbl                // label
)

const (

    // 如果设置, 剔除重复的函数声明.
    FilterFuncDuplicates MergeMode = 1 << iota

    // 如果设置, 剔除未关联 AST 节点的注释.
    FilterUnassociatedComments

    // 如果设置, 剔除重复导入声明.
    FilterImportDuplicates
)

const (
    SEND ChanDir = 1 << iota
    RECV
)

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.

// ArrayType 节点表示一个 array 或 slice 类型.
type ArrayType struct {
    Lbrack token.Pos // position of "["
    Len    Expr      // Ellipsis node for [...]T array types, nil for slice types
    Elt    Expr      // element type
}

// AssignStmt 节点表示一个赋值或短声明变量.
type AssignStmt struct {
    Lhs    []Expr
    TokPos token.Pos   // position of Tok
    Tok    token.Token // assignment token, DEFINE
    Rhs    []Expr
}

// A declaration is represented by one of the following declaration nodes.

// BadDecl 占位节点表示错误的声明语法或无法创建正确的声明节点.
type BadDecl struct {
    From, To token.Pos // position range of bad declaration
}

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.

// BadExpr 占位节点表示错误的表达式或无法创建正确的表达式节点.
type BadExpr struct {
    From, To token.Pos // position range of bad expression
}

// A statement is represented by a tree consisting of one
// or more of the following concrete statement nodes.

// BadStmt 占位节点表示错误的语句或无法创建正确的语句节点.
type BadStmt struct {
    From, To token.Pos // position range of bad statement
}

// BasicLit 节点表示一个基本类型字面量.
type BasicLit struct {
    ValuePos token.Pos   // literal position
    Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
    Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
}

// BinaryExpr 节点表示一个二进制表达式.
type BinaryExpr struct {
    X     Expr        // left operand
    OpPos token.Pos   // position of Op
    Op    token.Token // operator
    Y     Expr        // right operand
}

// BlockStmt 节点表示一个大括号语句块.
type BlockStmt struct {
    Lbrace token.Pos // position of "{"
    List   []Stmt
    Rbrace token.Pos // position of "}"
}

// BranchStmt 节点表示一个 break, continue, goto, fallthrough 语句.
type BranchStmt struct {
    TokPos token.Pos   // position of Tok
    Tok    token.Token // keyword token (BREAK, CONTINUE, GOTO, FALLTHROUGH)
    Label  *Ident      // label name; or nil
}

// CallExpr 节点表示一个表达式后面跟着一个参数列表.
type CallExpr struct {
    Fun      Expr      // function expression
    Lparen   token.Pos // position of "("
    Args     []Expr    // function arguments; or nil
    Ellipsis token.Pos // position of "...", if any
    Rparen   token.Pos // position of ")"
}

// CaseClause 表示 switch 语句的一个 case 表达式 (或 type).
type CaseClause struct {
    Case  token.Pos // position of "case" or "default" keyword
    List  []Expr    // list of expressions or types; nil means default case
    Colon token.Pos // position of ":"
    Body  []Stmt    // statement list; or nil
}

// The direction of a channel type is indicated by one
// of the following constants.

// channel 类型方向, 由下的面常数表示.
type ChanDir int

// ChanType 节点表示一个 channel 类型.
type ChanType struct {
    Begin token.Pos // position of "chan" keyword or "<-" (whichever comes first)
    Arrow token.Pos // position of "<-" (token.NoPos if there is no "<-")
    Dir   ChanDir   // channel direction
    Value Expr      // value type
}

// CommClause 节点表示 select 语句的一个 case.
type CommClause struct {
    Case  token.Pos // position of "case" or "default" keyword
    Comm  Stmt      // send or receive statement; nil means default case
    Colon token.Pos // position of ":"
    Body  []Stmt    // statement list; or nil
}

// A Comment node represents a single //-style or /*-style comment.

// Comment 节点表示一个 // 或 /* 风格注释.
type Comment struct {
    Slash token.Pos // position of "/" starting the comment
    Text  string    // comment text (excluding '\n' for //-style comments)
}

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.

// CommentGroup 表示一个注释序列, 之间没有其他的标记也没有空行.
type CommentGroup struct {
    List []*Comment // len(List) > 0
}

// A CommentMap maps an AST node to a list of comment groups
// associated with it. See NewCommentMap for a description of
// the association.

// CommentMap 映射一个 AST节点到与其关联的 CommentGroup 列表
// 关联声明详见 NewCommentMap.
type CommentMap map[Node][]*CommentGroup

// CompositeLit 节点表示一个组合字面量.
type CompositeLit struct {
    Type   Expr      // literal type; or nil
    Lbrace token.Pos // position of "{"
    Elts   []Expr    // list of composite elements; or nil
    Rbrace token.Pos // position of "}"
}

// All declaration nodes implement the Decl interface.

// 所有声明节点实现了 Decl 接口.
type Decl interface {
    Node
}

// DeclStmt 节点表示一个声明列表语句.
type DeclStmt struct {
    Decl Decl // *GenDecl with CONST, TYPE, or VAR token
}

// DeferStmt 节点表示一个 defer 语句.
type DeferStmt struct {
    Defer token.Pos // position of "defer" keyword
    Call  *CallExpr
}

// Ellipsis 节点表示参数类型列表 "..." 或数组类型的 "..." 长度.
type Ellipsis struct {
    Ellipsis token.Pos // position of "..."
    Elt      Expr      // ellipsis element type (parameter lists only); or nil
}

// EmptyStmt 节点表示一个空语句. "position" 是紧接分号之前的位置.
type EmptyStmt struct {
    Semicolon token.Pos // position of preceding ";"
}

// All expression nodes implement the Expr interface.

// 所有表达式节点实现了 Expr 接口.
type Expr interface {
    Node
}

// ExprStmt 节点表示一个语句中的(独立)表达式.
type ExprStmt struct {
    X Expr // expression
}

// A Field represents a Field declaration list in a struct type,
// a method list in an interface type, or a parameter/result declaration
// in a signature.

// Field 表示结构类型中的一个字段声明, 接口类型的一个方法, 或一个签名中的参数/返
// 回值.
type Field struct {
    Doc     *CommentGroup // associated documentation; or nil
    Names   []*Ident      // field/method/parameter names; or nil if anonymous field
    Type    Expr          // field/method/parameter type
    Tag     *BasicLit     // field tag; or nil
    Comment *CommentGroup // line comments; or nil
}

// A FieldFilter may be provided to Fprint to control the output.

// FieldFilter 可给 Fprint 提供输出控制.
type FieldFilter func(name string, value reflect.Value) bool

// A FieldList represents a list of Fields, enclosed by parentheses or braces.

// FieldList 表示一个用括号或大括号包围的字段列表.
type FieldList struct {
    Opening token.Pos // position of opening parenthesis/brace, if any
    List    []*Field  // field list; or nil
    Closing token.Pos // position of closing parenthesis/brace, if any
}

// A File node represents a Go source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes
// via Doc and Comment fields.

// File 节点表示一个 Go 源文件.
//
// Comments 列表包含源文件中顺序出场的所有注释,
// 其中包括来自文档其他节点指向的注释和注释字段.
type File struct {
    Doc        *CommentGroup   // associated documentation; or nil
    Package    token.Pos       // position of "package" keyword
    Name       *Ident          // package name
    Decls      []Decl          // top-level declarations; or nil
    Scope      *Scope          // package scope (this file only)
    Imports    []*ImportSpec   // imports in this file
    Unresolved []*Ident        // unresolved identifiers in this file
    Comments   []*CommentGroup // list of all comments in the source file
}

type Filter func(string) bool

// ForStmt 表示一个 for 语句.
type ForStmt struct {
    For  token.Pos // position of "for" keyword
    Init Stmt      // initialization statement; or nil
    Cond Expr      // condition; or nil
    Post Stmt      // post iteration statement; or nil
    Body *BlockStmt
}

// FuncDecl 节点表示一个函数声明.
type FuncDecl struct {
    Doc  *CommentGroup // associated documentation; or nil
    Recv *FieldList    // receiver (methods); or nil (functions)
    Name *Ident        // function/method name
    Type *FuncType     // function signature: parameters, results, and position of "func" keyword
    Body *BlockStmt    // function body; or nil (forward declaration)
}

// FuncLit 节点表示一个函数字面量.
type FuncLit struct {
    Type *FuncType  // function type
    Body *BlockStmt // function body
}

// FuncType 节点表示一个函数类型.
type FuncType struct {
    Func    token.Pos  // position of "func" keyword (token.NoPos if there is no "func")
    Params  *FieldList // (incoming) parameters; non-nil
    Results *FieldList // (outgoing) results; or nil
}

// GenDecl 节点 (通用声明节点) 表示一个导入, 常量, 类型或者变量声明.
// 一个有效的 Lparen 位置 (Lparen.Line > 0) 指示一个括号的声明.
// Tok 值和 Specs 元件类型关系:
//
//     token.IMPORT  *ImportSpec
//     token.CONST   *ValueSpec
//     token.TYPE    *TypeSpec
//     token.VAR     *ValueSpec
type GenDecl struct {
    Doc    *CommentGroup // associated documentation; or nil
    TokPos token.Pos     // position of Tok
    Tok    token.Token   // IMPORT, CONST, TYPE, VAR
    Lparen token.Pos     // position of '(', if any
    Specs  []Spec
    Rparen token.Pos // position of ')', if any
}

// GoStmt 节点表示一个 go 语句.
type GoStmt struct {
    Go   token.Pos // position of "go" keyword
    Call *CallExpr
}

// Ident 节点表示一个标识符.
type Ident struct {
    NamePos token.Pos // identifier position
    Name    string    // identifier name
    Obj     *Object   // denoted object; or nil
}

// IfStmt 节点表示一个 if 语句.
type IfStmt struct {
    If   token.Pos // position of "if" keyword
    Init Stmt      // initialization statement; or nil
    Cond Expr      // condition
    Body *BlockStmt
    Else Stmt // else branch; or nil
}

// ImportSpec 节点表示单个包导入.
type ImportSpec struct {
    Doc     *CommentGroup // associated documentation; or nil
    Name    *Ident        // local package name (including "."); or nil
    Path    *BasicLit     // import path
    Comment *CommentGroup // line comments; or nil
    EndPos  token.Pos     // end of spec (overrides Path.Pos if nonzero)
}

// An Importer resolves import paths to package Objects.
// The imports map records the packages already imported,
// indexed by package id (canonical import path).
// An Importer must determine the canonical import path and
// check the map to see if it is already present in the imports map.
// If so, the Importer can return the map entry.  Otherwise, the
// Importer should load the package data for the given path into
// a new *Object (pkg), record pkg in the imports map, and then
// return pkg.

// Importer 解析包对象导入路径.
// imports map 记录了已经导入的包, 以包 id 为索引 (规范化导入路径).
// Importer 必须确定规范化导入路径和检查映射, 看它是否已经存在于导入映射.
// 如果是, Importer 可返回映射入口. 否则, Importer 应加载指定路径包数据到
// 一个新 *Object (pkg), 在导入映射中记录 pkg, 然后返回 pkg.
type Importer func(imports map[string]*Object, path string) (pkg *Object, err error)

// IncDecStmt 节点表示一个加法或减法语句.
type IncDecStmt struct {
    X      Expr
    TokPos token.Pos   // position of Tok
    Tok    token.Token // INC or DEC
}

// IndexExpr 节点表示一个后跟索引的表达式.
type IndexExpr struct {
    X      Expr      // expression
    Lbrack token.Pos // position of "["
    Index  Expr      // index expression
    Rbrack token.Pos // position of "]"
}

// InterfaceType 节点表示一个接口类型.
type InterfaceType struct {
    Interface  token.Pos  // position of "interface" keyword
    Methods    *FieldList // list of methods
    Incomplete bool       // true if (source) methods are missing in the Methods list
}

// KeyValueExpr 节点表示组合字面值中的 (key : value) 对.
type KeyValueExpr struct {
    Key   Expr
    Colon token.Pos // position of ":"
    Value Expr
}

// LabeledStmt 节点表示一个标签语句.
type LabeledStmt struct {
    Label *Ident
    Colon token.Pos // position of ":"
    Stmt  Stmt
}

// MapType 节点表示一个 map 类型.
type MapType struct {
    Map   token.Pos // position of "map" keyword
    Key   Expr
    Value Expr
}

// The MergeMode flags control the behavior of MergePackageFiles.

// MergeMode 标志控制 MergePackageFiles 的行为.
type MergeMode uint

// All node types implement the Node interface.

// 所有的节点类型实现了 Node 接口.
type Node interface {
    Pos() token.Pos // position of first character belonging to the node
    End() token.Pos // position of first character immediately after the node
}

// ObjKind describes what an object represents.

// ObjKind 描述一个对象表示什么.
type ObjKind int

// An Object describes a named language entity such as a package, constant,
// type, variable, function (incl. methods), or label.
//
// The Data fields contains object-specific data:
//
//     Kind    Data type         Data value
//     Pkg    *types.Package    package scope
//     Con     int               iota for the respective declaration
//     Con     != nil            constant value
//     Typ     *Scope            (used as method scope during type checking - transient)

// Object 描述具名实体, 例如包, 常量, 类型, 变量, 函数(包括方法), 或者标签.
//
// Data 字段包含具体对象相关的数据:
//
//     Kind    Data type         Data value
//     Pkg    *types.Package    package scope
//     Con     int               iota for the respective declaration
//     Con     != nil            constant value
//     Typ     *Scope            (used as method scope during type checking - transient)
type Object struct {
    Kind ObjKind
    Name string      // declared name
    Decl interface{} // corresponding Field, XxxSpec, FuncDecl, LabeledStmt, AssignStmt, Scope; or nil
    Data interface{} // object-specific data; or nil
    Type interface{} // placeholder for type information; may be nil
}

// A Package node represents a set of source files
// collectively building a Go package.

// Package 节点表示源文件集合共同构建的 Go 包.
type Package struct {
    Name    string             // package name
    Scope   *Scope             // package scope across all files
    Imports map[string]*Object // map of package id -> package object
    Files   map[string]*File   // Go source files by filename
}

// ParenExpr 节点表示一个括号内的表达式.
type ParenExpr struct {
    Lparen token.Pos // position of "("
    X      Expr      // parenthesized expression
    Rparen token.Pos // position of ")"
}

// RangeStmt 表示一个带 range 从句的 for 语句.
type RangeStmt struct {
    For        token.Pos   // position of "for" keyword
    Key, Value Expr        // Key, Value may be nil
    TokPos     token.Pos   // position of Tok; invalid if Key == nil
    Tok        token.Token // ILLEGAL if Key == nil, ASSIGN, DEFINE
    X          Expr        // value to range over
    Body       *BlockStmt
}

// ReturnStmt 表示一个 return 语句.
type ReturnStmt struct {
    Return  token.Pos // position of "return" keyword
    Results []Expr    // result expressions; or nil
}

// A Scope maintains the set of named language entities declared
// in the scope and a link to the immediately surrounding (outer)
// scope.

// Scope 维护一定范围内和链接到外围的语言具名实体声明作用域.
type Scope struct {
    Outer   *Scope
    Objects map[string]*Object
}

// SelectStmt 节点表示一个 select 语句.
type SelectStmt struct {
    Select token.Pos  // position of "select" keyword
    Body   *BlockStmt // CommClauses only
}

// SelectorExpr 节点表示一个表达式后跟一个选择器.
type SelectorExpr struct {
    X   Expr   // expression
    Sel *Ident // field selector
}

// SendStmt 节点表示一个 send 语句( <- ).
type SendStmt struct {
    Chan  Expr
    Arrow token.Pos // position of "<-"
    Value Expr
}

// SliceExpr 节点表示一个表达式后跟切片索引.
type SliceExpr struct {
    X      Expr      // expression
    Lbrack token.Pos // position of "["
    Low    Expr      // begin of slice range; or nil
    High   Expr      // end of slice range; or nil
    Max    Expr      // maximum capacity of slice; or nil
    Slice3 bool      // true if 3-index slice (2 colons present)
    Rbrack token.Pos // position of "]"
}

// A Spec node represents a single (non-parenthesized) import,
// constant, type, or variable declaration.

// Spec 表示类型 *ImportSpec, *ValueSpec, 和 *TypeSpec 之一.
type Spec interface {
    Node
}

// StarExpr 节点表示一个形如 "*" Expression 的表达式.
// 语义上它可能是一个一元 "*" 表达式, 或指针类型.
type StarExpr struct {
    Star token.Pos // position of "*"
    X    Expr      // operand
}

// All statement nodes implement the Stmt interface.

// 所有语句节点实现了 Stmt 接口.
type Stmt interface {
    Node
}

// StructType 节点表示一个结构体类型.
type StructType struct {
    Struct     token.Pos  // position of "struct" keyword
    Fields     *FieldList // list of field declarations
    Incomplete bool       // true if (source) fields are missing in the Fields list
}

// SwitchStmt 节点表示一个 switch 语句表达式.
type SwitchStmt struct {
    Switch token.Pos  // position of "switch" keyword
    Init   Stmt       // initialization statement; or nil
    Tag    Expr       // tag expression; or nil
    Body   *BlockStmt // CaseClauses only
}

// TypeAssertExpr 节点表示一个表达式后跟类型断言.
type TypeAssertExpr struct {
    X      Expr      // expression
    Lparen token.Pos // position of "("
    Type   Expr      // asserted type; nil means type switch X.(type)
    Rparen token.Pos // position of ")"
}

// TypeSpec 节点表示一个类型声明 (TypeSpec 产生).
type TypeSpec struct {
    Doc     *CommentGroup // associated documentation; or nil
    Name    *Ident        // type name
    Type    Expr          // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
    Comment *CommentGroup // line comments; or nil
}

// TypeSwitchStmt 节点表示一个 switch 类型测试语句.
type TypeSwitchStmt struct {
    Switch token.Pos  // position of "switch" keyword
    Init   Stmt       // initialization statement; or nil
    Assign Stmt       // x := y.(type) or y.(type)
    Body   *BlockStmt // CaseClauses only
}

// UnrayExpr 节点表示一个一元表达式. 一元 "*" 表达式由 StarExpr 节点表示.
type UnaryExpr struct {
    OpPos token.Pos   // position of Op
    Op    token.Token // operator
    X     Expr        // operand
}

// ValueSpec 节点表示一个常量或变量声明 (ConstSpec 或 VarSpec 产生).
type ValueSpec struct {
    Doc     *CommentGroup // associated documentation; or nil
    Names   []*Ident      // value names (len(Names) > 0)
    Type    Expr          // value type; or nil
    Values  []Expr        // initial values; or nil
    Comment *CommentGroup // line comments; or nil
}

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).

// Visitor 的 Visit 方法被 Walk 调派游历的每个节点.
// 如果返回的 w 非 nil, Walk 调派 w 游历每个非 nil 子节点,
// 然后调用 w.Visit(nil).
type Visitor interface {
    Visit(node Node) (w Visitor)
}

// FileExports trims the AST for a Go source file in place such that
// only exported nodes remain: all top-level identifiers which are not exported
// and their associated information (such as type, initial value, or function
// body) are removed. Non-exported fields and methods of exported types are
// stripped. The File.Comments list is not changed.
//
// FileExports reports whether there are exported declarations.

// FileExports 缩减 Go 文件 AST, 只留下导出部分:
// 删除所有非导出 top-level 标识符和相关信息 (如 类型, 初始值或者函数主体).
// 剥离导出类型中的非导出属性和方法. File.Comments 不变.
//
// 如果有导出声明, 返回 true, 否则返回 false.
func FileExports(src *File) bool

// FilterDecl trims the AST for a Go declaration in place by removing
// all names (including struct field and interface method names, but
// not from parameter lists) that don't pass through the filter f.
//
// FilterDecl reports whether there are any declared names left after
// filtering.

// FilterDecl 缩减 Go 声明 AST, 通过删除所有名称未通过过滤器 f 的
// (包括结构体字段名和接口方法名, 但非来自参数列表).
//
// 如果有任何声明在过滤后被保留下来, 返回 true, 否则返回 false.
func FilterDecl(decl Decl, f Filter) bool

// FilterFile trims the AST for a Go file in place by removing all
// names from top-level declarations (including struct field and
// interface method names, but not from parameter lists) that don't
// pass through the filter f. If the declaration is empty afterwards,
// the declaration is removed from the AST. Import declarations are
// always removed. The File.Comments list is not changed.
//
// FilterFile reports whether there are any top-level declarations
// left after filtering.

// FilterFile 缩减 Go File AST, 通过删除所有名称未通过过滤器 f 的 top-level 声明
// (包括结构体字段名和接口方法名). 如果之后某个声明是空的, 也被删除.
// File.Comments 不变.
//
// 如果有任何声明在过滤后被保留下来, 返回 true, 否则返回 false.
func FilterFile(src *File, f Filter) bool

// FilterPackage trims the AST for a Go package in place by removing
// all names from top-level declarations (including struct field and
// interface method names, but not from parameter lists) that don't
// pass through the filter f. If the declaration is empty afterwards,
// the declaration is removed from the AST. The pkg.Files list is not
// changed, so that file names and top-level package comments don't get
// lost.
//
// FilterPackage reports whether there are any top-level declarations
// left after filtering.

// FilterPackage 缩减 Go Package AST, 通过删除所有名称未通过过滤器 f 的
// top-level 声明 (包括结构体字段名和接口方法名). 如果之后该声明是空的, 该声明从
// AST 删除. pkg.Files 列表不变, 所以文件名和 top-level 包注释不会丢失.
//
// 如果有任何 top-level 声明在过滤后被保留下来, 返回 true, 否则返回 false.
func FilterPackage(pkg *Package, f Filter) bool

// Fprint prints the (sub-)tree starting at AST node x to w.
// If fset != nil, position information is interpreted relative
// to that file set. Otherwise positions are printed as integer
// values (file set specific offsets).
//
// A non-nil FieldFilter f may be provided to control the output:
// struct fields for which f(fieldname, fieldvalue) is true are
// printed; all others are filtered from the output. Unexported
// struct fields are never printed.

// Fprint 打印 AST 节点(子)树 x 到 w.
// 如果 fset != nil, position 信息被解释为相对于该 file set.
// 否则 positions 被当作 interger 值打印 (file set 中的偏移量).
//
// 一个 non-nil FieldFilter f 可用来提供控制输出:
// 那些 f(fieldname, fieldvalue) 为 true 的结构体字段被打印, 其它被过滤掉.
// 非导出结构体字段决不被打印.
func Fprint(w io.Writer, fset *token.FileSet, x interface{}, f FieldFilter) (err error)

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).

// Inspect 遍历 AST,  深度优先顺序: 首先调用 f(node); node 不为 nil.
// 如果 f 返回 true, Inspect 用 f 递归调派所有 non-nil 子节点.
func Inspect(node Node, f func(Node) bool)

// IsExported reports whether name is an exported Go symbol
// (that is, whether it begins with an upper-case letter).

// IsExported 报告 name 是否是一个 Go 导出符号 (即是否以大写字符开始).
func IsExported(name string) bool

// MergePackageFiles creates a file AST by merging the ASTs of the
// files belonging to a package. The mode flags control merging behavior.

// MergePackageFiles 通过合并同属于一个包的文件 ASTs 创建一个 AST 文件.
// mode 标志控制合并行为.
func MergePackageFiles(pkg *Package, mode MergeMode) *File

// NewCommentMap creates a new comment map by associating comment groups
// of the comments list with the nodes of the AST specified by node.
//
// A comment group g is associated with a node n if:
//
//   - g starts on the same line as n ends
//   - g starts on the line immediately following n, and there is
//     at least one empty line after g and before the next node
//   - g starts before n and is not associated to the node before n
//     via the previous rules
//
// NewCommentMap tries to associate a comment group to the "largest"
// node possible: For instance, if the comment is a line comment
// trailing an assignment, the comment is associated with the entire
// assignment rather than just the last operand in the assignment.

// NewCommentMap 按 node 指定的 AST 节点,
// 通过关联注释组列表创建一个新的 CommentMap.
//
// 一个注释组 (CommentGroup) g 和一个节点 n 关联条件:
//
//     - g 启始位置和 n 的结束位置在同一行
//     - g 启始行紧跟 n 之后, 并且 g 之后下一个节点前至少有一个空行
//     - g 启始于 n 之前, 并且通过前面的规则未关联到 n 之前的节点
//
// NewCommentMap 试图关联注释组到 "最大" 节点的可能性:
// 用例, 如果一个行注释尾随一个赋值语句,
// 该注释与整个赋值语句相关联, 而不仅仅是赋值语句最后一个操作数.
func NewCommentMap(fset *token.FileSet, node Node, comments []*CommentGroup) CommentMap

// NewIdent creates a new Ident without position.
// Useful for ASTs generated by code other than the Go parser.

// NewIdent 新建一个不带位置的 Ident. 可用于 Go 解析器以外代码生成的 ASTs.
func NewIdent(name string) *Ident

// NewObj creates a new object of a given kind and name.

// NewObj 新建一个给定 kind 和 name 的对象.
func NewObj(kind ObjKind, name string) *Object

// NewPackage creates a new Package node from a set of File nodes. It resolves
// unresolved identifiers across files and updates each file's Unresolved list
// accordingly. If a non-nil importer and universe scope are provided, they are
// used to resolve identifiers not declared in any of the package files. Any
// remaining unresolved identifiers are reported as undeclared. If the files
// belong to different packages, one package name is selected and files with
// different package names are reported and then ignored.
// The result is a package node and a scanner.ErrorList if there were errors.

// NewPackage 从一组 File 节点新建一个 Package 节点.
// 它解析跨文件未明确标识符和更新每个文件的未明确的名单.
// 因此, 如果提供了 importer (非 nil) 和 universe 范围,
// 它们被用于解决任何包文件中未明确的标识符. 报告任何仍未明确的标识符.
// 如果文件属于不同的包, 选中一个包, 报告别的包文件, 然后忽略.
// 如果有错误, 返回结果是一个包节点和一个 scanner.ErrorList.
func NewPackage(fset *token.FileSet, files map[string]*File, importer Importer, universe *Scope) (*Package, error)

// NewScope creates a new scope nested in the outer scope.

// NewScope 新建一个嵌套于外围的作用域.
func NewScope(outer *Scope) *Scope

// NotNilFilter returns true for field values that are not nil;
// it returns false otherwise.

// NotNilFilter 返回 v 是否非 nil;
func NotNilFilter(_ string, v reflect.Value) bool

// PackageExports trims the AST for a Go package in place such that
// only exported nodes remain. The pkg.Files list is not changed, so that
// file names and top-level package comments don't get lost.
//
// PackageExports reports whether there are exported declarations;
// it returns false otherwise.

// PackageExports 缩减 Go Package AST, 只留下导出部分.
// pkg.Files 不变, 所以文件名和 top-level 包注释不会丢失.
//
// 如果有导出声明, 返回 true, 否则返回 false.
func PackageExports(pkg *Package) bool

// Print prints x to standard output, skipping nil fields.
// Print(fset, x) is the same as Fprint(os.Stdout, fset, x, NotNilFilter).

// Print 打印 x 到标准输出, 忽略 nil 字段.
// Print(fset, x) 等同 Fprint(os.Stdout, fset, x, NotNilFilter).
func Print(fset *token.FileSet, x interface{}) error

// SortImports sorts runs of consecutive import lines in import blocks in f. It
// also removes duplicate imports when it is possible to do so without data
// loss.

// SortImports 对 f import 块连续行导入行排序. 它会删除重复的导入.
func SortImports(fset *token.FileSet, f *File)

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).

// Walk 遍历 AST,  深度优先顺序: 首先调用 v.Visit(node); node 不为 nil.
// 如果 v.Visit(node) 返回值 w 非 nil, Walk 调派 w 游历每个 non-nil 子节点,
// 然后调用 w.Visit(nil).
func Walk(v Visitor, node Node)

func (*ArrayType) End() token.Pos

func (*ArrayType) Pos() token.Pos

func (*AssignStmt) End() token.Pos

func (*AssignStmt) Pos() token.Pos

func (*BadDecl) End() token.Pos

// Pos and End implementations for declaration nodes.

// Pos 和 End 实现节点声明.
func (*BadDecl) Pos() token.Pos

func (*BadExpr) End() token.Pos

// Pos and End implementations for expression/type nodes.

// Pos 和 End 实现 表达式/类型 节点.
func (*BadExpr) Pos() token.Pos

func (*BadStmt) End() token.Pos

// Pos and End implementations for statement nodes.

// Pos 和 End 现实语句节点.
func (*BadStmt) Pos() token.Pos

func (*BasicLit) End() token.Pos

func (*BasicLit) Pos() token.Pos

func (*BinaryExpr) End() token.Pos

func (*BinaryExpr) Pos() token.Pos

func (*BlockStmt) End() token.Pos

func (*BlockStmt) Pos() token.Pos

func (*BranchStmt) End() token.Pos

func (*BranchStmt) Pos() token.Pos

func (*CallExpr) End() token.Pos

func (*CallExpr) Pos() token.Pos

func (*CaseClause) End() token.Pos

func (*CaseClause) Pos() token.Pos

func (*ChanType) End() token.Pos

func (*ChanType) Pos() token.Pos

func (*CommClause) End() token.Pos

func (*CommClause) Pos() token.Pos

func (*Comment) End() token.Pos

func (*Comment) Pos() token.Pos

func (*CommentGroup) End() token.Pos

func (*CommentGroup) Pos() token.Pos

// Text returns the text of the comment.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed. Multiple empty lines are
// reduced to one, and trailing space on lines is trimmed. Unless the result
// is empty, it is newline-terminated.

// Text 返回注释文本. 注释标记 (//, /*, 和 */), 一行注释的首个空白, 开头和尾部的
// 空行被删除. 多个空行被简化为一行, 缩减掉行尾空白. 它以换行结束, 除非结果为空.
func (*CommentGroup) Text() string

func (*CompositeLit) End() token.Pos

func (*CompositeLit) Pos() token.Pos

func (*DeclStmt) End() token.Pos

func (*DeclStmt) Pos() token.Pos

func (*DeferStmt) End() token.Pos

func (*DeferStmt) Pos() token.Pos

func (*Ellipsis) End() token.Pos

func (*Ellipsis) Pos() token.Pos

func (*EmptyStmt) End() token.Pos

func (*EmptyStmt) Pos() token.Pos

func (*ExprStmt) End() token.Pos

func (*ExprStmt) Pos() token.Pos

func (*Field) End() token.Pos

func (*Field) Pos() token.Pos

func (*FieldList) End() token.Pos

// NumFields returns the number of (named and anonymous fields) in a FieldList.

// NumFields 返回 FieldList 的数量 (具名和匿名字段).
func (*FieldList) NumFields() int

func (*FieldList) Pos() token.Pos

func (*File) End() token.Pos

func (*File) Pos() token.Pos

func (*ForStmt) End() token.Pos

func (*ForStmt) Pos() token.Pos

func (*FuncDecl) End() token.Pos

func (*FuncDecl) Pos() token.Pos

func (*FuncLit) End() token.Pos

func (*FuncLit) Pos() token.Pos

func (*FuncType) End() token.Pos

func (*FuncType) Pos() token.Pos

func (*GenDecl) End() token.Pos

func (*GenDecl) Pos() token.Pos

func (*GoStmt) End() token.Pos

func (*GoStmt) Pos() token.Pos

func (*Ident) End() token.Pos

// IsExported reports whether id is an exported Go symbol
// (that is, whether it begins with an uppercase letter).

// IsExported 报告 id 是否是一个 Go 导出符号 (即是否以大写字符开始)
func (*Ident) IsExported() bool

func (*Ident) Pos() token.Pos

func (*Ident) String() string

func (*IfStmt) End() token.Pos

func (*IfStmt) Pos() token.Pos

func (*ImportSpec) End() token.Pos

// Pos and End implementations for spec nodes.

// Pos 和 End 实现节点规定.
func (*ImportSpec) Pos() token.Pos

func (*IncDecStmt) End() token.Pos

func (*IncDecStmt) Pos() token.Pos

func (*IndexExpr) End() token.Pos

func (*IndexExpr) Pos() token.Pos

func (*InterfaceType) End() token.Pos

func (*InterfaceType) Pos() token.Pos

func (*KeyValueExpr) End() token.Pos

func (*KeyValueExpr) Pos() token.Pos

func (*LabeledStmt) End() token.Pos

func (*LabeledStmt) Pos() token.Pos

func (*MapType) End() token.Pos

func (*MapType) Pos() token.Pos

// Pos computes the source position of the declaration of an object name.
// The result may be an invalid position if it cannot be computed
// (obj.Decl may be nil or not correct).

// Pos 计算对象名声明的源位置.
// 有可能不能计算而返回一个无效位置 (可能 obj.Decl 为 nil 或不正确).
func (*Object) Pos() token.Pos

func (*Package) End() token.Pos

func (*Package) Pos() token.Pos

func (*ParenExpr) End() token.Pos

func (*ParenExpr) Pos() token.Pos

func (*RangeStmt) End() token.Pos

func (*RangeStmt) Pos() token.Pos

func (*ReturnStmt) End() token.Pos

func (*ReturnStmt) Pos() token.Pos

// Insert attempts to insert a named object obj into the scope s.
// If the scope already contains an object alt with the same name,
// Insert leaves the scope unchanged and returns alt. Otherwise
// it inserts obj and returns nil.

// Insert 尝试插入一个具名对象 obj 到作用域 s. 如果作用域 s 已包含同名对象 alt,
// 作用域不变并返回 alt. 否则插入 obj 并返回 nil.
func (*Scope) Insert(obj *Object) (alt *Object)

// Lookup returns the object with the given name if it is
// found in scope s, otherwise it returns nil. Outer scopes
// are ignored.

// Lookup 返回给定名称的对象, 如果在作用域 s 中被找到的话, 否则返回 nil.
// 忽略 Outer 作用域.
func (*Scope) Lookup(name string) *Object

// Debugging support

// 调试支持
func (*Scope) String() string

func (*SelectStmt) End() token.Pos

func (*SelectStmt) Pos() token.Pos

func (*SelectorExpr) End() token.Pos

func (*SelectorExpr) Pos() token.Pos

func (*SendStmt) End() token.Pos

func (*SendStmt) Pos() token.Pos

func (*SliceExpr) End() token.Pos

func (*SliceExpr) Pos() token.Pos

func (*StarExpr) End() token.Pos

func (*StarExpr) Pos() token.Pos

func (*StructType) End() token.Pos

func (*StructType) Pos() token.Pos

func (*SwitchStmt) End() token.Pos

func (*SwitchStmt) Pos() token.Pos

func (*TypeAssertExpr) End() token.Pos

func (*TypeAssertExpr) Pos() token.Pos

func (*TypeSpec) End() token.Pos

func (*TypeSpec) Pos() token.Pos

func (*TypeSwitchStmt) End() token.Pos

func (*TypeSwitchStmt) Pos() token.Pos

func (*UnaryExpr) End() token.Pos

func (*UnaryExpr) Pos() token.Pos

func (*ValueSpec) End() token.Pos

func (*ValueSpec) Pos() token.Pos

// Comments returns the list of comment groups in the comment map.
// The result is sorted is source order.

// Comments 返回 CommentMap 中的注释组列表. 返回值排序为源码中的顺序.
func (CommentMap) Comments() []*CommentGroup

// Filter returns a new comment map consisting of only those
// entries of cmap for which a corresponding node exists in
// the AST specified by node.

// Filter 返回一个新的 CommentMap, 由 cmap 中与指定的 AST 节点相关项组成.
func (CommentMap) Filter(node Node) CommentMap

func (CommentMap) String() string

// Update replaces an old node in the comment map with the new node
// and returns the new node. Comments that were associated with the
// old node are associated with the new node.

// Update 用新节点替换 cmap 中的旧节点并返回新节点.
// 与旧节点关联的注释与新节点关联.
func (CommentMap) Update(old, new Node) Node

func (ObjKind) String() string

