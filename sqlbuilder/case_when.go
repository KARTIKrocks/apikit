package sqlbuilder

import "strings"

// CaseBuilder builds SQL CASE expressions.
type CaseBuilder struct {
	operand    string
	hasOperand bool
	whens      []whenClause
	elseSQL    string
	elseArgs   []any
	hasElse    bool
}

type whenClause struct {
	condSQL  string
	condArgs []any
	thenSQL  string
	thenArgs []any
}

// Case creates a new CaseBuilder. If an operand is provided, it becomes
// a simple CASE (CASE operand WHEN ...); otherwise it's a searched CASE.
func Case(operand ...string) *CaseBuilder {
	cb := &CaseBuilder{}
	if len(operand) > 0 {
		cb.operand = operand[0]
		cb.hasOperand = true
	}
	return cb
}

// When adds a WHEN/THEN pair with literal SQL.
func (cb *CaseBuilder) When(cond, result string) *CaseBuilder {
	cb.whens = append(cb.whens, whenClause{condSQL: cond, thenSQL: result})
	return cb
}

// WhenExpr adds a parameterized WHEN/THEN pair.
func (cb *CaseBuilder) WhenExpr(condSQL string, condArgs []any, thenSQL string, thenArgs []any) *CaseBuilder {
	cb.whens = append(cb.whens, whenClause{
		condSQL:  condSQL,
		condArgs: condArgs,
		thenSQL:  thenSQL,
		thenArgs: thenArgs,
	})
	return cb
}

// Else sets the ELSE clause with a literal SQL string.
func (cb *CaseBuilder) Else(result string) *CaseBuilder {
	cb.elseSQL = result
	cb.hasElse = true
	return cb
}

// ElseExpr sets the ELSE clause with a parameterized expression.
func (cb *CaseBuilder) ElseExpr(result string, args ...any) *CaseBuilder {
	cb.elseSQL = result
	cb.elseArgs = args
	cb.hasElse = true
	return cb
}

// End builds the CASE expression and returns it as an Expr.
func (cb *CaseBuilder) End() Expr {
	var sb strings.Builder
	sb.Grow(64)
	var allArgs []any
	offset := 0

	sb.WriteString("CASE")
	if cb.hasOperand {
		sb.WriteByte(' ')
		sb.WriteString(cb.operand)
	}

	for _, w := range cb.whens {
		sb.WriteString(" WHEN ")
		rebased := rebasePlaceholders(w.condSQL, offset)
		sb.WriteString(rebased)
		allArgs = append(allArgs, w.condArgs...)
		offset += len(w.condArgs)

		sb.WriteString(" THEN ")
		rebased = rebasePlaceholders(w.thenSQL, offset)
		sb.WriteString(rebased)
		allArgs = append(allArgs, w.thenArgs...)
		offset += len(w.thenArgs)
	}

	if cb.hasElse {
		sb.WriteString(" ELSE ")
		rebased := rebasePlaceholders(cb.elseSQL, offset)
		sb.WriteString(rebased)
		allArgs = append(allArgs, cb.elseArgs...)
	}

	sb.WriteString(" END")
	return Expr{SQL: sb.String(), Args: allArgs}
}
