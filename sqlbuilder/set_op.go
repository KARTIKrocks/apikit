package sqlbuilder

// setOpKind is the type of set operation.
type setOpKind int

const (
	setOpUnion setOpKind = iota
	setOpUnionAll
	setOpIntersect
	setOpExcept
)

// setOp represents a set operation (UNION, INTERSECT, EXCEPT) with another query.
type setOp struct {
	kind  setOpKind
	query Query
}

// setOpKeyword returns the SQL keyword for a set operation.
func setOpKeyword(k setOpKind) string {
	switch k {
	case setOpUnion:
		return "UNION"
	case setOpUnionAll:
		return "UNION ALL"
	case setOpIntersect:
		return "INTERSECT"
	case setOpExcept:
		return "EXCEPT"
	default:
		return "UNION"
	}
}
