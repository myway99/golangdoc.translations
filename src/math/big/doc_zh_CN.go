// Copyright The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

// Package big implements arbitrary-precision arithmetic (big numbers).
// The following numeric types are supported:
//
//     Int    signed integers
//     Rat    rational numbers
//     Float  floating-point numbers
//
// The zero value for an Int, Rat, or Float correspond to 0. Thus, new
// values can be declared in the usual ways and denote 0 without further
// initialization:
//
//     var x Int        // &x is an *Int of value 0
//     var r = &Rat{}   // r is a *Rat of value 0
//     y := new(Float)  // y is a *Float of value 0
//
// Alternatively, new values can be allocated and initialized with factory
// functions of the form:
//
//     func NewT(v V) *T
//
// For instance, NewInt(x) returns an *Int set to the value of the int64
// argument x, NewRat(a, b) returns a *Rat set to the fraction a/b where
// a and b are int64 values, and NewFloat(f) returns a *Float initialized
// to the float64 argument f. More flexibility is provided with explicit
// setters, for instance:
//
//     var z1 Int
//     z1.SetUint64(123)                 // z1 := 123
//     z2 := new(Rat).SetFloat64(1.2)    // z2 := 6/5
//     z3 := new(Float).SetInt(z1)       // z3 := 123.0
//
// Setters, numeric operations and predicates are represented as methods of
// the form:
//
//     func (z *T) SetV(v V) *T          // z = v
//     func (z *T) Unary(x *T) *T        // z = unary x
//     func (z *T) Binary(x, y *T) *T    // z = x binary y
//     func (x *T) Pred() P              // p = pred(x)
//
// with T one of Int, Rat, or Float. For unary and binary operations, the
// result is the receiver (usually named z in that case; see below); if it
// is one of the operands x or y it may be safely overwritten (and its memory
// reused).
//
// Arithmetic expressions are typically written as a sequence of individual
// method calls, with each call corresponding to an operation. The receiver
// denotes the result and the method arguments are the operation's operands.
// For instance, given three *Int values a, b and c, the invocation
//
//     c.Add(a, b)
//
// computes the sum a + b and stores the result in c, overwriting whatever
// value was held in c before. Unless specified otherwise, operations permit
// aliasing of parameters, so it is perfectly ok to write
//
//     sum.Add(sum, x)
//
// to accumulate values x in a sum.
//
// (By always passing in a result value via the receiver, memory use can be
// much better controlled. Instead of having to allocate new memory for each
// result, an operation can reuse the space allocated for the result value,
// and overwrite that value with the new result in the process.)
//
// Notational convention: Incoming method parameters (including the receiver)
// are named consistently in the API to clarify their use. Incoming operands
// are usually named x, y, a, b, and so on, but never z. A parameter specifying
// the result is named z (typically the receiver).
//
// For instance, the arguments for (*Int).Add are named x and y, and because
// the receiver specifies the result destination, it is called z:
//
//     func (z *Int) Add(x, y *Int) *Int
//
// Methods of this form typically return the incoming receiver as well, to
// enable simple call chaining.
//
// Methods which don't require a result value to be passed in (for instance,
// Int.Sign), simply return the result. In this case, the receiver is typically
// the first operand, named x:
//
//     func (x *Int) Sign() int
//
// Various methods support conversions between strings and corresponding
// numeric values, and vice versa: *Int, *Rat, and *Float values implement
// the Stringer interface for a (default) string representation of the value,
// but also provide SetString methods to initialize a value from a string in
// a variety of supported formats (see the respective SetString documentation).
//
// Finally, *Int, *Rat, and *Float satisfy the fmt package's Scanner interface
// for scanning and (except for *Rat) the Formatter interface for formatted
// printing.

// big 包实现了（大数的）高精度运算. 它支持以下数值类型：
//
//     - Int    带符号整数
//     - Rat    有理数
//
// 典型的方法形式如下：
//
//     func (z *Int) Op(x, y *Int) *Int    （*Rat 同理）
//
// 它实现了像 z = x Op y 这样的操作，并将其结果作为接收者；若接收者为操作数之一
// ， 其值可能会被覆盖（而内存则会被重用）。为保留操作，其结果也会被返回。若该方
// 法返回除 *Int 或 *Rat 之外的结果，其中一个操作数将被作为接收者。
package big

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    "math"
    "math/rand"
    "strconv"
    "strings"
    "sync"
)

// MaxBase is the largest number base accepted for string conversions.
const MaxBase = 'z' - 'a' + 10 + 1 // = hexValue('z') + 1


// An Int represents a signed multi-precision integer.
// The zero value for an Int represents the value 0.

// Int 表示一个带符号多精度整数。 Int 的零值为值 0。
type Int struct {
}

// A Rat represents a quotient a/b of arbitrary precision. The zero value for a
// Rat represents the value 0.
type Rat struct {
}

// A Word represents a single digit of a multi-precision unsigned integer.

// Word 表示多精度无符号整数的单个数字。
type Word uintptr

// NewInt allocates and returns a new Int set to x.

// NewInt 为 x 分配并返回一个新的 Int。
func NewInt(x int64) *Int

// NewRat creates a new Rat with numerator a and denominator b.
func NewRat(a, b int64) *Rat

// Abs sets z to |x| (the absolute value of x) and returns z.

// Abs 将 z 置为 |x|（即 x 的绝对值）并返回 z。
func (*Int) Abs(x *Int) *Int

// Add sets z to the sum x+y and returns z.

// Add 将 z 置为 x+y 的和并返回 z。
func (*Int) Add(x, y *Int) *Int

// And sets z = x & y and returns z.

// And 置 z = x & y 并返回 z。
func (*Int) And(x, y *Int) *Int

// AndNot sets z = x &^ y and returns z.

// AndNot 置 z = x &^ y 并返回 z。
func (*Int) AndNot(x, y *Int) *Int

// Binomial sets z to the binomial coefficient of (n, k) and returns z.

// Binomial 将 z 置为 (n, k) 的二项式系数并返回 z。
func (*Int) Binomial(n, k int64) *Int

// Bit returns the value of the i'th bit of x. That is, it
// returns (x>>i)&1. The bit index i must be >= 0.

// Bit 返回 x 第 i 位的值。换言之，它返回 (x>>i)&1。位下标 i 必须 >= 0。
func (*Int) Bit(i int) uint

// BitLen returns the length of the absolute value of x in bits.
// The bit length of 0 is 0.

// BitLen 返回 z 的绝对值的位数长度。0 的位长为 0.
func (*Int) BitLen() int

// Bits provides raw (unchecked but fast) access to x by returning its
// absolute value as a little-endian Word slice. The result and x share
// the same underlying array.
// Bits is intended to support implementation of missing low-level Int
// functionality outside this package; it should be avoided otherwise.

// Bits 提供了对 z 的原始访问（未经检查但很快）。它通过将其绝对值作为小端序的
// Word 切片返回来实现。其结果与 x 共享同一底层数组。Bits 旨在支持此包外缺失的底
// 层 Int 功能的实现，除此之外应尽量避免。
func (*Int) Bits() []Word

// Bytes returns the absolute value of x as a big-endian byte slice.

// Bytes 将 x 的绝对值作为大端序的字节切片返回。
func (*Int) Bytes() []byte

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y

// Cmp 比较 x 和 y 并返回：
//
//     若 x <  y 则为 -1
//     若 x == y 则为  0
//     若 x >  y 则为 +1
func (*Int) Cmp(y *Int) (r int)

// Div sets z to the quotient x/y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
// Div implements Euclidean division (unlike Go); see DivMod for more details.

// Div 在 y != 0 时，将 z 置为 x/y 的商并返回 z。 若 y ==
// 0，就会产生一个除以零的运行时派错。 Div
// 实现了欧氏除法（与Go不同），更多详情见 DivMod。
func (*Int) Div(x, y *Int) *Int

// DivMod sets z to the quotient x div y and m to the modulus x mod y
// and returns the pair (z, m) for y != 0.
// If y == 0, a division-by-zero run-time panic occurs.
//
// DivMod implements Euclidean division and modulus (unlike Go):
//
//     q = x div y  such that
//     m = x - y*q  with 0 <= m < |y|
//
// (See Raymond T. Boute, ``The Euclidean definition of the functions
// div and mod''. ACM Transactions on Programming Languages and
// Systems (TOPLAS), 14(2):127-144, New York, NY, USA, 4/1992.
// ACM press.)
// See QuoRem for T-division and modulus (like Go).

// DivMod 在 y != 0 时，将 z 置为 x 除以 y 的商，将 m 置为 x 取模 y 的模数并返回
// 值对 (z, m)。 若 y == 0，就会产生一个除以零的运行时派错。
//
// DivMod 实现了截断式除法和取模（与Go不同）：
//
//     q = x div y // 使得
//     m = x - y*q // 其中
//     0 <= m < |q|
//
// （详见 Raymond T. Boute，《函数 div 和 mod 的欧氏定义》以及《ACM编程语言与系
// 统会议记录》 （TOPLAS），14(2):127-144, New York, NY, USA, 4/1992. ACM 出版社
// 。） 截断式除法和取模（与Go相同）见 QuoRem。
func (*Int) DivMod(x, y, m *Int) (*Int, *Int)

// Exp sets z = x**y mod |m| (i.e. the sign of m is ignored), and returns z.
// If y <= 0, the result is 1 mod |m|; if m == nil or m == 0, z = x**y.
// See Knuth, volume 2, section 4.6.3.

// Exp 置 z = x**y mod |m|（换言之，m 的符号被忽略），并返回 z。 若 y <=0，则其
// 结果为 1，若 m == nil 或 m == 0，则 z = x**y。 见 Knuth《计算机程序设计艺术》
// ，卷 2，章节 4.6.3。
func (*Int) Exp(x, y, m *Int) *Int

// Format is a support routine for fmt.Formatter. It accepts
// the formats 'b' (binary), 'o' (octal), 'd' (decimal), 'x'
// (lowercase hexadecimal), and 'X' (uppercase hexadecimal).
// Also supported are the full suite of package fmt's format
// verbs for integral types, including '+', '-', and ' '
// for sign control, '#' for leading zero in octal and for
// hexadecimal, a leading "0x" or "0X" for "%#x" and "%#X"
// respectively, specification of minimum digits precision,
// output field width, space or zero padding, and left or
// right justification.

// Format 是 fmt.Formatter 的一个支持函数。它接受 'b'（二进制）、'o'（八进制）、
// 'd'（十进制）、'x'（小写十六进制）和 'X'（大写十六进制）的格式。也同样支持
// fmt 包的一整套类型的格式占位符，包括用于符号控制的 '+'、'-' 和 ' '，用于八进
// 制前导零的 '#'，分别用于十六进制前导 "0x" 或 "0X" 的 "%#x" 和 "%#X"，用于最小
// 数字精度的规范， 输出字段的宽度，空格或零的填充，以及左右对齐。
func (*Int) Format(s fmt.State, ch rune)

// GCD sets z to the greatest common divisor of a and b, which both must
// be > 0, and returns z.
// If x and y are not nil, GCD sets x and y such that z = a*x + b*y.
// If either a or b is <= 0, GCD sets z = x = y = 0.

// GCD 将 z 置为 a 和 b 的最大公约数，二者必须均 > 0，并返回 z。 若 x 或 y 非
// nil，GCD 会设置 x 与 y 的值使得 z = a*x + b*y。 若 a 或 b <= 0，GCD就会置 z =
// x = y = 0。
func (*Int) GCD(x, y, a, b *Int) *Int

// GobDecode implements the gob.GobDecoder interface.

// GobDecode 实现了 gob.GobDecoder 接口。
func (*Int) GobDecode(buf []byte) error

// GobEncode implements the gob.GobEncoder interface.

// GobEncode 实现了 gob.GobEncoder 接口。
func (*Int) GobEncode() ([]byte, error)

// Int64 returns the int64 representation of x.
// If x cannot be represented in an int64, the result is undefined.

// Int64 返回 x 的 int64 表示。 若 x 不能被表示为
// int64，则其结果是未定义的。
func (*Int) Int64() int64

// Lsh sets z = x << n and returns z.

// Lsh 置 z = x << n 并返回 z。
func (*Int) Lsh(x *Int, n uint) *Int

// MarshalJSON implements the json.Marshaler interface.

// MarshalJSON 实现了 json.Marshaler 接口。
func (*Int) MarshalJSON() ([]byte, error)

// MarshalText implements the encoding.TextMarshaler interface.
func (*Int) MarshalText() (text []byte, err error)

// Mod sets z to the modulus x%y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
// Mod implements Euclidean modulus (unlike Go); see DivMod for more details.

// Mod 在 y != 0 时，将 z 置为 x%y 的余数并返回 z。 若 y ==
// 0，就会产生一个除以零的运行时派错。 Mod
// 实现了欧氏取模（与Go不同），更多详情见 DivMod。
func (*Int) Mod(x, y *Int) *Int

// ModInverse sets z to the multiplicative inverse of g in the ring ℤ/nℤ
// and returns z. If g and n are not relatively prime, the result is undefined.

// ModInverse 将 z 置为 g 在环 ℤ/nℤ 中的乘法逆元素并返回 z。若 g 与 n
// 并不互质，则结果为未定义。
func (*Int) ModInverse(g, n *Int) *Int

// Mul sets z to the product x*y and returns z.

// Mul 将 z 置为 x*y 的积并返回 z。
func (*Int) Mul(x, y *Int) *Int

// MulRange sets z to the product of all integers
// in the range [a, b] inclusively and returns z.
// If a > b (empty range), the result is 1.

// MulRange 将 z 置为闭区间 [a, b] 内所有整数的积并返回 z。 若 a >
// b（空区间），则其结果为 1。
func (*Int) MulRange(a, b int64) *Int

// Neg sets z to -x and returns z.

// Neg 将 z 置为 -x 并返回 z。
func (*Int) Neg(x *Int) *Int

// Not sets z = ^x and returns z.

// Not 置 z = ^x 并返回 z。
func (*Int) Not(x *Int) *Int

// Or sets z = x | y and returns z.

// Or 置 z = x | y 并返回 z。
func (*Int) Or(x, y *Int) *Int

// ProbablyPrime performs n Miller-Rabin tests to check whether x is prime.
// If x is prime, it returns true.
// If x is not prime, it returns false with probability at least 1 - ¼ⁿ.
//
// It is not suitable for judging primes that an adversary may have crafted
// to fool this test.

// ProbablyPrime 通过执行 n 次 Miller-Rabin 测试来检查 x 是否为质数。 若它返回
// true，x 有 1 - 1/4^n 的可能性为质数。 若它返回 false，则 x 不是质数。x 必须
// >0。
func (*Int) ProbablyPrime(n int) bool

// Quo sets z to the quotient x/y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
// Quo implements truncated division (like Go); see QuoRem for more details.

// Quo 在 y != 0 时，将 z 置为 x/y 的商并返回 z。 若 y ==
// 0，就会产生一个除以零的运行时派错。 Quo
// 实现了截断式除法（与Go相同），更多详情见 QuoRem。
func (*Int) Quo(x, y *Int) *Int

// QuoRem sets z to the quotient x/y and r to the remainder x%y
// and returns the pair (z, r) for y != 0.
// If y == 0, a division-by-zero run-time panic occurs.
//
// QuoRem implements T-division and modulus (like Go):
//
//     q = x/y      with the result truncated to zero
//     r = x - y*q
//
// (See Daan Leijen, ``Division and Modulus for Computer Scientists''.)
// See DivMod for Euclidean division and modulus (unlike Go).

// QuoRem 在 y != 0 时，将 z 置为 x/y 的商，将 r 置为 x%y 的余数并返回值对 (z,
// r)。 若 y == 0，就会产生一个除以零的运行时派错。
//
// QuoRem 实现了截断式除法和取模（与Go相同）：
//
//     q = x/y      // 其结果向零截断
//     r = x - y*q
//
// （详见 Daan Leijen，《计算机科学家的除法和取模》。） 欧氏除法和取模（与Go不同
// ）见 DivMod。
func (*Int) QuoRem(x, y, r *Int) (*Int, *Int)

// Rand sets z to a pseudo-random number in [0, n) and returns z.

// Rand 将 z 置为区间 [0, n) 中的一个伪随机数并返回 z。
func (*Int) Rand(rnd *rand.Rand, n *Int) *Int

// Rem sets z to the remainder x%y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
// Rem implements truncated modulus (like Go); see QuoRem for more details.

// Rem 在 y != 0 时，将 z 置为 x%y 的余数并返回 z。 若 y ==
// 0，就会产生一个除以零的运行时派错。 Rem
// 实现了截断式取模（与Go相同），更多详情见 QuoRem。
func (*Int) Rem(x, y *Int) *Int

// Rsh sets z = x >> n and returns z.

// Rsh 置 z = x >> n 并返回 z。
func (*Int) Rsh(x *Int, n uint) *Int

// Scan is a support routine for fmt.Scanner; it sets z to the value of
// the scanned number. It accepts the formats 'b' (binary), 'o' (octal),
// 'd' (decimal), 'x' (lowercase hexadecimal), and 'X' (uppercase hexadecimal).

// Scan 是 fmt.Scanner 的一个支持函数；它将 z
// 置为已扫描数字的值。它接受格式'b'（二进制）、
// 'o'（八进制）、'd'（十进制）、'x'（小写十六进制）及'X'（大写十六进制）。
func (*Int) Scan(s fmt.ScanState, ch rune) error

// Set sets z to x and returns z.

// Set 将 z 置为 x 并返回 z。
func (*Int) Set(x *Int) *Int

// SetBit sets z to x, with x's i'th bit set to b (0 or 1).
// That is, if b is 1 SetBit sets z = x | (1 << i);
// if b is 0 SetBit sets z = x &^ (1 << i). If b is not 0 or 1,
// SetBit will panic.

// SetBit 将 z 置为 x，将 x 的第 i 位置为 b（0 或 1）。 换言之，若 b 为 1，
// SetBit 会置 z = x | (1 << i)；若 b 为 0，SetBit 会置 z = x &^ (1 << i)。若 b
// 非 0 或 1，SetBit 就会引发派错。
func (*Int) SetBit(x *Int, i int, b uint) *Int

// SetBits provides raw (unchecked but fast) access to z by setting its
// value to abs, interpreted as a little-endian Word slice, and returning
// z. The result and abs share the same underlying array.
// SetBits is intended to support implementation of missing low-level Int
// functionality outside this package; it should be avoided otherwise.

// SetBits 提供了对 z 的原始访问（未经检查但很快）。它通过将其值设为 abs，解释为
// 小端序的 Word 切片，并返回 z 来实现。SetBits 旨在支持此包外缺失的底层 Int 功
// 能的实现，除此之外应尽量避免。
func (*Int) SetBits(abs []Word) *Int

// SetBytes interprets buf as the bytes of a big-endian unsigned
// integer, sets z to that value, and returns z.

// SetBytes 将 buf 解释为大端序的无符号整数字节，置 z 为该值后返回 z。
func (*Int) SetBytes(buf []byte) *Int

// SetInt64 sets z to x and returns z.

// SetInt64 将 z 置为 x 并返回 z。
func (*Int) SetInt64(x int64) *Int

// SetString sets z to the value of s, interpreted in the given base,
// and returns z and a boolean indicating success. If SetString fails,
// the value of z is undefined but the returned value is nil.
//
// The base argument must be 0 or a value between 2 and MaxBase. If the base
// is 0, the string prefix determines the actual conversion base. A prefix of
// ``0x'' or ``0X'' selects base 16; the ``0'' prefix selects base 8, and a
// ``0b'' or ``0B'' prefix selects base 2. Otherwise the selected base is 10.

// SetString 将 z 置为 s 的值，按给定的进制 base 解释并返回 z 和一个指示是否成功
// 的布尔值。 若 SetString 失败，则 z 的值是未定义的，其返回值则为 nil。
//
// 进制实参 base 必须为 0 或从 2 到 MaxBase 的值。若 base 为 0，则其实际的转换进
// 制由 该字符串的前缀决定。前缀“0x”或“0X”会选择16进制，前缀“0”会选择8进制
// ，前缀“0b”或“0B” 会选择2进制。其它情况则选择10进制。
func (*Int) SetString(s string, base int) (*Int, bool)

// SetUint64 sets z to x and returns z.

// SetUint64 将 z 置为 x 并返回 z。
func (*Int) SetUint64(x uint64) *Int

// Sign returns:
//
//     -1 if x <  0
//      0 if x == 0
//     +1 if x >  0

// 符号返回：
//
//     若 x <  0 则为 -1
//     若 x == 0 则为  0
//     若 x >  0 则为 +1
func (*Int) Sign() int

func (*Int) String() string

// Sub sets z to the difference x-y and returns z.

// Sub 将 z 置为 x-y 的差并返回 z。
func (*Int) Sub(x, y *Int) *Int

// Uint64 returns the uint64 representation of x.
// If x cannot be represented in a uint64, the result is undefined.

// Uint64 返回 x 的 uint64 表示。 若 x 不能被表示为
// uint64，则其结果是未定义的。
func (*Int) Uint64() uint64

// UnmarshalJSON implements the json.Unmarshaler interface.

// UnmarshalJSON 实现了 json.Unmarshaler 接口。
func (*Int) UnmarshalJSON(text []byte) error

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (*Int) UnmarshalText(text []byte) error

// Xor sets z = x ^ y and returns z.

// Xor 置 z = x ^ y 并返回 z。
func (*Int) Xor(x, y *Int) *Int

// Abs sets z to |x| (the absolute value of x) and returns z.
func (*Rat) Abs(x *Rat) *Rat

// Add sets z to the sum x+y and returns z.
func (*Rat) Add(x, y *Rat) *Rat

// Cmp compares x and y and returns:
//
//     -1 if x <  y
//      0 if x == y
//     +1 if x >  y
func (*Rat) Cmp(y *Rat) int

// Denom returns the denominator of x; it is always > 0. The result is a
// reference to x's denominator; it may change if a new value is assigned to x,
// and vice versa.
func (*Rat) Denom() *Int

// Float32 returns the nearest float32 value for x and a bool indicating whether
// f represents x exactly. If the magnitude of x is too large to be represented
// by a float32, f is an infinity and exact is false. The sign of f always
// matches the sign of x, even if f == 0.
func (*Rat) Float32() (f float32, exact bool)

// Float64 returns the nearest float64 value for x and a bool indicating whether
// f represents x exactly. If the magnitude of x is too large to be represented
// by a float64, f is an infinity and exact is false. The sign of f always
// matches the sign of x, even if f == 0.
func (*Rat) Float64() (f float64, exact bool)

// FloatString returns a string representation of x in decimal form with prec
// digits of precision after the decimal point. The last digit is rounded to
// nearest, with halves rounded away from zero.

// FloatString returns a string representation of x in decimal form with prec
// digits of precision after the decimal point and the last digit rounded.
func (*Rat) FloatString(prec int) string

// GobDecode implements the gob.GobDecoder interface.
func (*Rat) GobDecode(buf []byte) error

// GobEncode implements the gob.GobEncoder interface.
func (*Rat) GobEncode() ([]byte, error)

// Inv sets z to 1/x and returns z.
func (*Rat) Inv(x *Rat) *Rat

// IsInt reports whether the denominator of x is 1.

// IsInt returns true if the denominator of x is 1.
func (*Rat) IsInt() bool

// MarshalText implements the encoding.TextMarshaler interface.
func (*Rat) MarshalText() (text []byte, err error)

// Mul sets z to the product x*y and returns z.
func (*Rat) Mul(x, y *Rat) *Rat

// Neg sets z to -x and returns z.
func (*Rat) Neg(x *Rat) *Rat

// Num returns the numerator of x; it may be <= 0. The result is a reference to
// x's numerator; it may change if a new value is assigned to x, and vice versa.
// The sign of the numerator corresponds to the sign of x.
func (*Rat) Num() *Int

// Quo sets z to the quotient x/y and returns z. If y == 0, a division-by-zero
// run-time panic occurs.
func (*Rat) Quo(x, y *Rat) *Rat

// RatString returns a string representation of x in the form "a/b" if b != 1,
// and in the form "a" if b == 1.
func (*Rat) RatString() string

// Scan is a support routine for fmt.Scanner. It accepts the formats 'e', 'E',
// 'f', 'F', 'g', 'G', and 'v'. All formats are equivalent.
func (*Rat) Scan(s fmt.ScanState, ch rune) error

// Set sets z to x (by making a copy of x) and returns z.
func (*Rat) Set(x *Rat) *Rat

// SetFloat64 sets z to exactly f and returns z. If f is not finite, SetFloat
// returns nil.
func (*Rat) SetFloat64(f float64) *Rat

// SetFrac sets z to a/b and returns z.
func (*Rat) SetFrac(a, b *Int) *Rat

// SetFrac64 sets z to a/b and returns z.
func (*Rat) SetFrac64(a, b int64) *Rat

// SetInt sets z to x (by making a copy of x) and returns z.
func (*Rat) SetInt(x *Int) *Rat

// SetInt64 sets z to x and returns z.
func (*Rat) SetInt64(x int64) *Rat

// SetString sets z to the value of s and returns z and a boolean indicating
// success. s can be given as a fraction "a/b" or as a floating-point number
// optionally followed by an exponent. If the operation failed, the value of z
// is undefined but the returned value is nil.
func (*Rat) SetString(s string) (*Rat, bool)

// Sign returns:
//
//     -1 if x <  0
//      0 if x == 0
//     +1 if x >  0
func (*Rat) Sign() int

// String returns a string representation of x in the form "a/b" (even if b ==
// 1).
func (*Rat) String() string

// Sub sets z to the difference x-y and returns z.
func (*Rat) Sub(x, y *Rat) *Rat

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (*Rat) UnmarshalText(text []byte) error

