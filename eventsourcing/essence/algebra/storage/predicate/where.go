package predicate

type Where struct {
	Predicate Predicate
	Params    ParamBinds
}

func MustQuery(query string, params ParamBinds) *Where {
	predicates, err := Parse(query)
	if err != nil {
		panic(err)
	}

	return &Where{
		Predicate: predicates,
		Params:    params,
	}
}
