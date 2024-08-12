package z

// Pair is two element tuple.
type Pair[T1, T2 any] struct {
	First  T1
	Second T2
}

// MakePair creates a new Pair.
func MakePair[T1, T2 any](k T1, v T2) Pair[T1, T2] {
	return Pair[T1, T2]{
		First:  k,
		Second: v,
	}
}

// Triple is three element tuple.
type Triple[T1, T2, T3 any] struct {
	First  T1
	Second T2
	Third  T3
}

// MakeTriple creates a new Triple.
func MakeTriple[T1, T2, T3 any](v1 T1, v2 T2, v3 T3) Triple[T1, T2, T3] {
	return Triple[T1, T2, T3]{
		First:  v1,
		Second: v2,
		Third:  v3,
	}
}
