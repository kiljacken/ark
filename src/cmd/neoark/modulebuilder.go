package main

import (
	"github.com/ark-lang/ark/src/parser"
)

type ModuleBuilder struct {
	moduleLookup *parser.ModuleLookup
	depGraph     *parser.DependencyGraph
}

func NewModuleBuilder() *ModuleBuilder {
	res := &ModuleBuilder{}
	res.moduleLookup = parser.NewModuleLookup("")
	res.depGraph = parser.NewDependencyGraph()
	return res
}

func (m *ModuleBuilder) DoDirectory(dirpath string) {

}

func (m *ModuleBuilder) DoFile(filepath string) {

}
