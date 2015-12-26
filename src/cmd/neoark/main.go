package main

import (
	"flag"
	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/semantic"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("Expected 1 argument, got %d", flag.NArg())
	}

	argument := flag.Arg(0)
	log.Printf("Got argument: %s", argument)

	if isDir(argument) {
		log.Printf("Argument was directory")
		doModule(argument)
	} else if isArkSource(argument) {
		log.Printf("Argument was ark source file")
		doFreestanding(argument)
	} else {
		log.Fatalf("Argument was neither a directory or an ark source file")
	}

	/*
		builder := NewModuleBuilder()
		builder.AddFile(filepath)

		if builder.HasCyclicDependency() {
			log.Printf("Encountered cyclic dependency: %s",
					   builder.DescribeCyclicDependency())
		}

		builder.Construct()
	*/
}

func doFreestanding(filepath string) {
	sourcefile, err := lexer.NewSourcefile(filepath)
	if err != nil {
		log.Fatalf("Encountered error reading source file: %s", err)
	}

	log.Printf("Lexing file '%s'", sourcefile.Name)
	tokens := lexer.Lex(sourcefile)

	log.Printf("Parsing file '%s'", sourcefile.Name)
	parsetree, deps := parser.Parse(sourcefile)

	// TODO: We really shouldn't have to do this here
	parsetree.Name = sourcefile.Name

	// TODO: Handle dependencies

	log.Printf("Constructing file '%s'", sourcefile.Name)

	_, _, _ = tokens, parsetree, deps
}

func doModule(dirpath string) {

}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func isArkSource(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return strings.HasSuffix(fi.Name(), ".ark")
}
