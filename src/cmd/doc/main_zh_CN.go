// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

// Doc (usually run as go doc) accepts zero, one or two arguments.
//
// Zero arguments:
//     go doc
// Show the documentation for the package in the current directory.
//
// One argument:
//     go doc <pkg>
//     go doc <sym>[.<method>]
//     go doc [<pkg>.]<sym>[.<method>]
//     go doc [<pkg>.][<sym>.]<method>
// The first item in this list that succeeds is the one whose documentation
// is printed. If there is a symbol but no package, the package in the current
// directory is chosen. However, if the argument begins with a capital
// letter it is always assumed to be a symbol in the current directory.
//
// Two arguments:
//     go doc <pkg> <sym>[.<method>]
//
// Show the documentation for the package, symbol, and method. The
// first argument must be a full package path. This is similar to the
// command-line usage for the godoc command.
//
// For commands, unless the -cmd flag is present "go doc command"
// shows only the package-level docs for the package.
//
// For complete documentation, run "go help doc".
package main // go get cmd/doc

import (
    "bytes"
    "flag"
    "fmt"
    "go/ast"
    "go/build"
    "go/doc"
    "go/format"
    "go/parser"
    "go/token"
    "io"
    "log"
    "os"
    "path"
    "path/filepath"
    "regexp"
    "runtime"
    "strings"
    "testing"
    "unicode"
    "unicode/utf8"
)

// Dirs is a structure for scanning the directory tree.
// Its Next method returns the next Go source directory it finds.
// Although it can be used to scan the tree multiple times, it
// only walks the tree once, caching the data it finds.
type Dirs struct {
    scan   chan string // directories generated by walk.
    paths  []string    // Cache of known paths.
    offset int         // Counter for Next.
}

type Package struct {
    writer     io.Writer // Destination for output.
    name       string    // Package name, json for encoding/json.
    userPath   string    // String the user used to find this package.
    unexported bool
    matchCase  bool
    pkg        *ast.Package // Parsed package.
    file       *ast.File    // Merged from all files in the package
    doc        *doc.Package
    build      *build.Package
    fs         *token.FileSet // Needed for printing.
    buf        bytes.Buffer
}

type PackageError string // type returned by pkg.Fatalf.


func TestDoc(t *testing.T)

// Test the code to try multiple packages. Our test case is
//     go doc rand.Float64
// This needs to find math/rand.Float64; however crypto/rand, which doesn't
// have the symbol, usually appears first in the directory listing.
func TestMultiplePackages(t *testing.T)

func TestTrim(t *testing.T)

// Next returns the next directory in the scan. The boolean
// is false when the scan is done.
func (*Dirs) Next() (string, bool)

// Reset puts the scan back at the beginning.
func (*Dirs) Reset()

// pkg.Fatalf is like log.Fatalf, but panics so it can be recovered in the
// main do function, so it doesn't cause an exit. Allows testing to work
// without running a subprocess. The log prefix will be added when
// logged in main; it is not added here.
func (*Package) Fatalf(format string, args ...interface{})

func (*Package) Printf(format string, args ...interface{})

func (PackageError) Error() string

