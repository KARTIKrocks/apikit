package sqlbuilder

// cte represents a Common Table Expression (WITH clause).
type cte struct {
	name      string
	query     Query
	recursive bool
}
