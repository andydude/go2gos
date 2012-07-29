
package main

import (
//	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime/pprof"
)

const debug = false

func usage() {
}

func compile(filename string) {
	// open input file
	rd, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// open output file
	//wr, err := os.Create(filename + "s")
	//if err != nil {
	//	  panic(err)
	//}

	// find guile
	guile, err := exec.LookPath("guile")
	if err != nil {
		panic(err)
	}

	// pretty-print
	const pretty = "(begin (use-modules (ice-9 pretty-print)) (pretty-print (read)))"
	cmd := exec.Command(guile, "-c", pretty)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	pr, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	// compile to pipe
	c := NewCompiler()
	c.Compile(rd, pr)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	rd.Close()
	//wr.Close()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Errorf("ERROR: you must give exactly one command-line argument.")
		usage()
	}
	filename := os.Args[1]

	if debug {
		out, err := os.Create("profile")
		if err != nil { panic(err) }
		err = pprof.StartCPUProfile(out)
		if err != nil { panic(err) }
		compile(filename)
		pprof.StopCPUProfile()
		out.Close()
	} else {
		compile(filename)
	}
}