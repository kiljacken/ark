package parser

import (
	"fmt"
	"github.com/ark-lang/ark/src/util/log"
)

type TypeCategory int

const (
	CAT_INTEGER TypeCategory = iota
	CAT_NUMBER
	CAT_POINTER
)

func (v TypeCategory) String() string {
	switch v {
	case CAT_INTEGER:
		return "integer"
	case CAT_NUMBER:
		return "number"
	case CAT_POINTER:
		return "pointer"
	}
	panic("Invalid type category")
}

type AnnotatedExpr struct {
	Expr Expr
	Id   int
}

type TypeConstraint interface {
	Substitute(a, b int)
	String() string
	Equals(other TypeConstraint) bool
	Indirection() int
	MainId() int
}

type EqualsConstraint struct{ A, B int }
type PointerConstraint struct{ A, B int }
type DerefConstraint struct{ A, B int }
type IsCategoryConstraint struct {
	Id       int
	Category TypeCategory
}
type IsConstraint struct {
	Id   int
	Type Type
}

func (v *EqualsConstraint) Substitute(a, b int) {
	if v.A == a {
		v.A = b
	}
	if v.B == a {
		v.B = b
	}
}

func (v *PointerConstraint) Substitute(a, b int) {
	if v.A == a {
		v.A = b
	}
	if v.B == a {
		v.B = b
	}
}

func (v *DerefConstraint) Substitute(a, b int) {
	if v.A == a {
		v.A = b
	}
	if v.B == a {
		v.B = b
	}
}

func (v *IsCategoryConstraint) Substitute(a, b int) {
	if v.Id == a {
		v.Id = b
	}
}

func (v *IsConstraint) Substitute(a, b int) {
	if v.Id == a {
		v.Id = b
	}
}

func (v *EqualsConstraint) String() string {
	return fmt.Sprintf("$%d = $%d", v.A, v.B)
}

func (v *PointerConstraint) String() string {
	return fmt.Sprintf("$%d is &$%d", v.A, v.B)
}

func (v *DerefConstraint) String() string {
	return fmt.Sprintf("$%d is ^$%d", v.A, v.B)
}

func (v *IsCategoryConstraint) String() string {
	return fmt.Sprintf("$%d is %v", v.Id, v.Category)
}

func (v *IsConstraint) String() string {
	return fmt.Sprintf("$%d is `%s`", v.Id, v.Type.TypeName())
}

func (v *EqualsConstraint) Equals(other TypeConstraint) bool {
	oec, ok := other.(*EqualsConstraint)
	return ok && v.A == oec.A && v.B == oec.B
}

func (v *PointerConstraint) Equals(other TypeConstraint) bool {
	opc, ok := other.(*PointerConstraint)
	return ok && v.A == opc.A && v.B == opc.B
}

func (v *DerefConstraint) Equals(other TypeConstraint) bool {
	odc, ok := other.(*DerefConstraint)
	return ok && v.A == odc.A && v.B == odc.B
}

func (v *IsCategoryConstraint) Equals(other TypeConstraint) bool {
	oc, ok := other.(*IsCategoryConstraint)
	return ok && v.Id == oc.Id && v.Category == oc.Category
}

func (v *IsConstraint) Equals(other TypeConstraint) bool {
	oic, ok := other.(*IsConstraint)
	return ok && v.Id == oic.Id && v.Type.Equals(oic.Type)
}

func (v *EqualsConstraint) Indirection() int {
	// Value doesn't matter
	return -1
}

func (v *PointerConstraint) Indirection() int {
	return 1
}

func (v *DerefConstraint) Indirection() int {
	return 1
}

func (v *IsCategoryConstraint) Indirection() int {
	return 2
}

func (v *IsConstraint) Indirection() int {
	return 0
}

func (v *EqualsConstraint) MainId() int {
	return v.A
}
func (v *PointerConstraint) MainId() int {
	return v.A
}
func (v *DerefConstraint) MainId() int {
	return v.A
}
func (v *IsCategoryConstraint) MainId() int {
	return v.Id
}
func (v *IsConstraint) MainId() int {
	return v.Id
}

type NewInferer struct {
	Module      *Module
	Function    *Function
	Exprs       []*AnnotatedExpr
	Constraints []TypeConstraint
	IdCount     int
}

func NewInferer_(mod *Module) *NewInferer {
	return &NewInferer{
		Module: mod,
	}
}

func (v *NewInferer) AddConstraint(c TypeConstraint) {
	v.Constraints = append(v.Constraints, c)
	//v.ConstraintMap[c.MainId()] = append(v.ConstraintMap[c.MainId()], c)
}

func (v *NewInferer) AddEqualsConstraint(a, b int) {
	v.AddConstraint(&EqualsConstraint{A: a, B: b})
}

func (v *NewInferer) AddPointerConstraint(a, b int) {
	v.AddConstraint(&PointerConstraint{A: a, B: b})
}

func (v *NewInferer) AddDerefConstraint(a, b int) {
	v.AddConstraint(&DerefConstraint{A: a, B: b})
}

func (v *NewInferer) AddIsCategoryConstraint(id int, cat TypeCategory) {
	v.AddConstraint(&IsCategoryConstraint{Id: id, Category: cat})
}

func (v *NewInferer) AddIsConstraint(id int, typ Type) {
	v.AddConstraint(&IsConstraint{Id: id, Type: typ})
}

func (v *NewInferer) CullNils() {
	newConstraints := make([]TypeConstraint, 0, len(v.Constraints))
	for _, cons := range v.Constraints {
		if cons != nil {
			newConstraints = append(newConstraints, cons)
		}
	}
	v.Constraints = newConstraints
}

func (v *NewInferer) SubstituteEquals() {
	for idx, c := range v.Constraints {
		if eqc, ok := c.(*EqualsConstraint); ok {
			v.Constraints[idx] = nil

			for _, cons := range v.Constraints {
				if cons != nil {
					cons.Substitute(eqc.A, eqc.B)
				}
			}

			for _, expr := range v.Exprs {
				if expr.Id == eqc.A {
					expr.Id = eqc.B
				}
			}
		}
	}

	v.CullNils()
}

func (v *NewInferer) RemoveDuplicates() {
	for oi, c := range v.Constraints {
		if c == nil {
			continue
		}

		for ii, ic := range v.Constraints {
			if oi == ii || ic == nil {
				continue
			}

			if c.Equals(ic) {
				v.Constraints[ii] = nil
			}
		}
	}

	v.CullNils()
}

func (v *NewInferer) PickMostSpecific() {
	for _, c := range v.Constraints {
		if c == nil {
			continue
		}
		for ii, ic := range v.Constraints {
			if ic == nil || ic.MainId() != c.MainId() {
				continue
			}
			if ic.Indirection() > c.Indirection() {
				v.Constraints[ii] = nil
			}
		}
	}

	v.CullNils()
}

func (v *NewInferer) EnterScope(s *Scope) {}

func (v *NewInferer) ExitScope(s *Scope) {
	if s == nil {
		log.Debugln("inferer", "=== END OF SCOPE ===")
		for _, ann := range v.Exprs {
			log.Debugln("inferer", "Type variable `$%d` bound to `%s`", ann.Id, ann.Expr)
			log.Debugln("inferer", "%s", v.Module.File.MarkSpan(ann.Expr.Pos()))
		}

		log.Debugln("inferer", "=== BEFORE SUBSTITUTION ===")
		log.Debugln("inferer", "%d constraints", len(v.Constraints))
		for _, con := range v.Constraints {
			log.Debugln("inferer", "Constraint: %s", con.String())
		}

		v.SubstituteEquals()
		log.Debugln("inferer", "=== AFTER SUBSTITUTION ===")
		log.Debugln("inferer", "%d constraints", len(v.Constraints))
		for _, con := range v.Constraints {
			log.Debugln("inferer", "Constraint: %s", con.String())
		}

		v.RemoveDuplicates()
		log.Debugln("inferer", "=== AFTER DEDUPLICATION ===")
		log.Debugln("inferer", "%d constraints", len(v.Constraints))
		for _, con := range v.Constraints {
			log.Debugln("inferer", "Constraint: %s", con.String())
		}

		v.PickMostSpecific()
		log.Debugln("inferer", "=== AFTER SPECIALIZATION ===")
		log.Debugln("inferer", "%d constraints", len(v.Constraints))
		for _, con := range v.Constraints {
			log.Debugln("inferer", "Constraint: %s", con.String())
		}
	}
}

func (v *NewInferer) PostVisit(n Node) {
	switch n.(type) {
	case *FunctionDecl:
		// TODO: Nested functions
		v.Function = nil
		return
	}
}

func (v *NewInferer) Visit(n Node) {
	switch n.(type) {
	case *FunctionDecl:
		// TODO: Nested functions
		v.Function = n.(*FunctionDecl).Function
		return
	}

	switch n.(type) {
	case *VariableDecl:
		vd := n.(*VariableDecl)
		if vd.Assignment != nil {
			v.HandleExpr(vd.Assignment)
		}

	case *AssignStat:
		as := n.(*AssignStat)
		a := v.HandleExpr(as.Access)
		b := v.HandleExpr(as.Assignment)
		v.AddEqualsConstraint(a, b)

	case *BinopAssignStat:
		bas := n.(*BinopAssignStat)
		a := v.HandleExpr(bas.Access)
		b := v.HandleExpr(bas.Assignment)
		v.AddEqualsConstraint(a, b)

	case *CallStat:
		cs := n.(*CallStat)
		v.HandleExpr(cs.Call)

	case *DeferStat:
		ds := n.(*DeferStat)
		v.HandleExpr(ds.Call)

	case *IfStat:
		is := n.(*IfStat)
		for _, expr := range is.Exprs {
			id := v.HandleExpr(expr)
			v.AddIsConstraint(id, PRIMITIVE_bool)
		}

	case *ReturnStat:
		rs := n.(*ReturnStat)
		if rs.Value != nil {
			id := v.HandleExpr(rs.Value)
			v.AddIsConstraint(id, v.Function.ReturnType)
		}

	case *LoopStat:
		ls := n.(*LoopStat)
		if ls.Condition != nil {
			id := v.HandleExpr(ls.Condition)
			v.AddIsConstraint(id, PRIMITIVE_bool)
		}

	case *MatchStat:
		// todo
	}
}

func (v *NewInferer) HandleExpr(expr Expr) int {
	log.Debugln("inferer", "=== NEW EXPR ===")
	log.Debugln("inferer", "Expr is type `%T`", expr)
	log.Debugln("inferer", "Expr has value `%v`\n", expr)
	log.Debugln("inferer", "%s", v.Module.File.MarkSpan(expr.Pos()))

	ann := &AnnotatedExpr{Id: v.IdCount, Expr: expr}
	v.Exprs = append(v.Exprs, ann)
	v.IdCount++

	switch expr.(type) {
	case *BinaryExpr:
		be := expr.(*BinaryExpr)
		a := v.HandleExpr(be.Lhand)
		b := v.HandleExpr(be.Rhand)
		switch be.Op {
		case BINOP_EQ, BINOP_NOT_EQ, BINOP_GREATER, BINOP_LESS,
			BINOP_GREATER_EQ, BINOP_LESS_EQ:
			v.AddEqualsConstraint(a, b)
			v.AddIsCategoryConstraint(a, CAT_NUMBER)
			v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

		case BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR:
			v.AddEqualsConstraint(a, b)
			v.AddIsCategoryConstraint(a, CAT_INTEGER)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD:
			// TODO: These assumptions don't hold once we add operator overloading
			v.AddEqualsConstraint(a, b)
			v.AddIsCategoryConstraint(a, CAT_NUMBER)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_BIT_LEFT, BINOP_BIT_RIGHT:
			v.AddIsCategoryConstraint(a, CAT_INTEGER)
			v.AddIsCategoryConstraint(b, CAT_INTEGER)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_LOG_AND, BINOP_LOG_OR:
			v.AddIsConstraint(a, PRIMITIVE_bool)
			v.AddIsConstraint(b, PRIMITIVE_bool)
			v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

		default:
			panic("Unhandled binary operator in type inference")

		}

	case *UnaryExpr:
		ue := expr.(*UnaryExpr)
		id := v.HandleExpr(ue.Expr)
		switch ue.Op {
		case UNOP_LOG_NOT:
			v.AddIsConstraint(id, PRIMITIVE_bool)
			v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

		case UNOP_BIT_NOT:
			v.AddIsCategoryConstraint(id, CAT_INTEGER)
			v.AddEqualsConstraint(ann.Id, id)

		case UNOP_NEGATIVE:
			v.AddIsCategoryConstraint(id, CAT_NUMBER)
			v.AddEqualsConstraint(ann.Id, id)

		case UNOP_DEREF:
			v.AddIsCategoryConstraint(id, CAT_POINTER)
			v.AddDerefConstraint(ann.Id, id)

		}

	case *CallExpr:
		ce := expr.(*CallExpr)
		if ce.Function.ReturnType != nil {
			v.AddIsConstraint(ann.Id, ce.Function.ReturnType)
		} else {
			log.Debugln("inferer", "Function return type was nil, assuming `void`")
			v.AddIsConstraint(ann.Id, PRIMITIVE_void)
		}

		for idx, arg := range ce.Arguments {
			id := v.HandleExpr(arg)

			if idx < len(ce.Function.Parameters) {
				// We're not in a vararg
				param := ce.Function.Parameters[idx]
				v.AddIsConstraint(id, param.Variable.Type)
			}
		}

	case *CastExpr:
		ce := expr.(*CastExpr)
		id := v.HandleExpr(ce.Expr)
		_ = id
		v.AddIsConstraint(ann.Id, ce.Type)

	case *AddressOfExpr:
		aoe := expr.(*AddressOfExpr)
		id := v.HandleExpr(aoe.Access)
		v.AddPointerConstraint(ann.Id, id)
		v.AddIsCategoryConstraint(ann.Id, CAT_POINTER)

	case *VariableAccessExpr:
		vae := expr.(*VariableAccessExpr)
		if vae.Variable.Type != nil {
			v.AddIsConstraint(ann.Id, vae.Variable.Type)
		}

	case *StructAccessExpr:
		sae := expr.(*StructAccessExpr)
		if sae.Variable.Type != nil {
			v.AddIsConstraint(ann.Id, sae.Variable.Type)
		} else {
			log.Debugln("inferer", "Struct access expr had nil type")
		}

	case *BoolLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

	case *NumericLiteral:
		v.AddIsCategoryConstraint(ann.Id, CAT_NUMBER)

	case *StringLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_str)

	case *RuneLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_rune)

	default:
		log.Debugln("inferer", "Unhandled expression type `%T`", expr)
	}

	return ann.Id
}
