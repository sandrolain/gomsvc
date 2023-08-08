package control

func PanicWithError[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}

func PanicIfError(e error) {
	if e != nil {
		panic(e)
	}
}
