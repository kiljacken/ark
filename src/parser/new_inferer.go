package parser

import (
	"fmt"
	"github.com/ark-lang/ark/src/util"
	"github.com/ark-lang/ark/src/util/log"
)

type TypeVariable struct {
	MetaType
	Id int
}

func (v *TypeVariable) Equals(other Type) bool {
	if ot, ok := other.(*TypeVariable); ok {
		return v.Id == ot.Id
	}
	return false
}

func (v *TypeVariable) String() string {
	return "(" + util.Blue("TypeVariable") + ": " + v.TypeName() + ")"
}

func (v *TypeVariable) TypeName() string {
	return fmt.Sprintf("$%d", v.Id)
}

func (v *TypeVariable) ActualType() Type {
	return v
}

func (v *TypeVariable) resolveType(src Locatable, res *Resolver, s *Scope) Type {
	panic("TypeVariable encountered in resolve")
}

type AnnotatedExpr struct {
	Expr Expr
	Id   int
}

type SideType int

const (
	IdentSide SideType = iota
	TypeSide
)

type Side struct {
	Id       int
	SideType SideType
	Type     Type
}

func (v *Side) Subs(id int, left *Side) {
	switch v.SideType {
	case IdentSide:
		if v.Id == id {
			*v = *left
		}

	case TypeSide:
		v.Type

	default:
		panic("Invalid SideType")
	}
}

func SubsType(typ Type, id int, left *Side) Type {
	visited := make(map[Type]bool)

	for {
		switch typ {
		case *NamedType:
			nt := typ.(*NamedType)
			nt.Type = SubsType(nt.Type, id, left)

		case PointerType:
			pt := typ.(PointerType)
			pt.Addressee = SubsType(pt.Addressee, id, left)

		case ArrayType:
			at := typ.(ArrayType)
		case *StructType:
			st := typ.(*StructType)
		case *TupleType:
			tt := typ.(*TupleType)
		case *TypeVariable:
		}
	}

	return typ
}

func (v *Side) String() string {
	switch v.SideType {
	case IdentSide:
		return fmt.Sprintf("$%d", v.Id)
	case TypeSide:
		return fmt.Sprintf("type `%s`", v.Type.TypeName())
	}
	panic("Invalid side type")
}

type Constraint struct {
	Left, Right *Side
}

func (v *Constraint) String() string {
	return fmt.Sprintf("%s = %s", v.Left, v.Right)
}

type Substitution struct {
	Constraint
}

type NewInferer struct {
	Module      *Module
	Function    *Function
	Exprs       []*AnnotatedExpr
	Constraints []*Constraint
	IdCount     int
}

func NewInferer_(mod *Module) *NewInferer {
	return &NewInferer{
		Module: mod,
	}
}

func (v *NewInferer) AddConstraint(c *Constraint) {
	v.Constraints = append(v.Constraints, c)
}

func (v *NewInferer) AddEqualsConstraint(a, b int) {
	c := &Constraint{
		Left:  &Side{Id: a, SideType: IdentSide},
		Right: &Side{Id: b, SideType: IdentSide},
	}
	v.AddConstraint(c)
}

func (v *NewInferer) AddIsConstraint(id int, typ Type) {
	c := &Constraint{
		Left:  &Side{Id: id, SideType: IdentSide},
		Right: &Side{Type: typ, SideType: TypeSide},
	}
	v.AddConstraint(c)
}

func (v *NewInferer) Unify() []*Substitution {
	stack := make([]*Constraint, len(v.Constraints))
	copy(stack, v.Constraints)

	var substitutions []*Substitution

	for len(stack) > 0 {
		idx := len(stack) - 1

		element := stack[idx]
		stack = stack[:idx]
		left, right := element.Left, element.Right

		log.Debugln("inferer", "\nThe constraint: %s", element)

		// 1. If X and Y are identical identifiers, do nothing.
		if left.SideType == right.SideType {
			var equal bool
			switch left.SideType {
			case IdentSide:
				equal = (left.Id == right.Id)

			case TypeSide:
				equal = (left.Type.Equals(right.Type))
			}

			if equal {
				log.Debugln("inferer", "Case 1")
				continue
			}
		}

		// 2. If X is an identifier, replace all occurrences of X by Y both on the stack and in the substitution, and
		// add X → Y to the substitution.
		if left.SideType == IdentSide {
			log.Debugln("inferer", "Case 2")
			for _, cons := range stack {
				cons.Left.Subs(left.Id, right)
				cons.Right.Subs(left.Id, right)
			}
			for _, cons := range substitutions {
				cons.Left.Subs(left.Id, right)
				cons.Right.Subs(left.Id, right)
			}
			substitutions = append(substitutions, &Substitution{Constraint: *element})
			continue
		}

		// 3. If Y is an identifier, replace all occurrences of Y by X both on the stack and in the substitution, and
		// add Y → X to the substitution.
		if right.SideType == IdentSide {
			log.Debugln("inferer", "Case 3")
			for _, cons := range stack {
				cons.Left.Subs(right.Id, left)
				cons.Right.Subs(right.Id, left)
			}
			for _, cons := range substitutions {
				cons.Left.Subs(right.Id, left)
				cons.Right.Subs(right.Id, left)
			}
			substitutions = append(substitutions, &Substitution{Constraint: *element})
			continue
		}

		// 4. If X is of the form C(X_1, ..., X_n) for some constructor C, and Y is of the form C(Y_1, ..., Y_n) (i.e., it
		// has the same constructor), then push X_i = Y_i for all 1 ≤ i ≤ n onto the stack.

		// 5. Otherwise, X and Y do not unify. Report an error.
		log.Errorln("inferer", "Constraint did not unify. This is an error (%s)", element)
	}

	return substitutions
}

func (v *NewInferer) EnterScope(s *Scope) {}

func (v *NewInferer) ExitScope(s *Scope) {
	if s == nil {
		log.Debugln("inferer", "=== END OF SCOPE ===")
		for _, ann := range v.Exprs {
			log.Debugln("inferer", "Type variable `$%d` bound to `%s`", ann.Id, ann.Expr)
			log.Debugln("inferer", "%s", v.Module.File.MarkSpan(ann.Expr.Pos()))
		}

		log.Debugln("inferer", "=== CONSTRAINTS ===")
		for _, cons := range v.Constraints {
			log.Debugln("inferer", "Constraint: %s", cons)
		}

		substitutions := v.Unify()
		log.Debugln("inferer", "=== SUBSTITUTIONS ===")
		for _, subs := range substitutions {
			log.Debugln("inferer", "Substitution: %s", subs)
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
			v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

		case BINOP_BIT_AND, BINOP_BIT_OR, BINOP_BIT_XOR:
			v.AddEqualsConstraint(a, b)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_ADD, BINOP_SUB, BINOP_MUL, BINOP_DIV, BINOP_MOD:
			// TODO: These assumptions don't hold once we add operator overloading
			v.AddEqualsConstraint(a, b)
			v.AddEqualsConstraint(ann.Id, a)

		case BINOP_BIT_LEFT, BINOP_BIT_RIGHT:
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
			v.AddEqualsConstraint(ann.Id, id)

		case UNOP_NEGATIVE:
			v.AddEqualsConstraint(ann.Id, id)

		case UNOP_DEREF:
			// TODO: v.AddDerefConstraint(ann.Id, id)

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
		v.AddIsConstraint(ann.Id, pointerTo(&TypeVariable{Id: id}))

	case *SizeofExpr:
		soe := expr.(*SizeofExpr)
		if soe.Expr != nil {
			id := v.HandleExpr(soe.Expr)
			_ = id
		}
		v.AddIsConstraint(ann.Id, PRIMITIVE_uint)

	case *DefaultExpr:
		de := expr.(*DefaultExpr)
		v.AddIsConstraint(ann.Id, de.Type)

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

	case *ArrayAccessExpr:
		aae := expr.(*ArrayAccessExpr)
		if aae.Array.GetType() != nil {
			v.AddIsConstraint(ann.Id, aae.Array.GetType().ActualType().(ArrayType).MemberType)
		} else {
			log.Debugln("inferer", "Struct access expr had nil type")
		}

	/*case *DerefAccessExpr:
	dae := expr.(*DerefAccessExpr)
	_ = dae.Expr*/

	case *EnumLiteral:
		// TODO: Infer type via constructor
		el := expr.(*EnumLiteral)
		if el.Type != nil {
			v.AddIsConstraint(ann.Id, el.Type)
		} else {
			log.Debugln("inferer", "Enum literal had nil type")
		}

	case *BoolLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_bool)

	case *NumericLiteral:
		// TODO: Set type here?

	case *StringLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_str)

	case *RuneLiteral:
		v.AddIsConstraint(ann.Id, PRIMITIVE_rune)

	default:
		log.Errorln("inferer", "Unhandled expression type `%T`", expr)
	}

	return ann.Id
}

/*func (v *NewInferer) CullNils() {
	newConstraints := make([]*Constraint, 0, len(v.Constraints))
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
}*/
