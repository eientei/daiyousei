package check

func Check0(err error) {
	if err != nil {
		panic(err)
	}
}

func Check1(out interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return out
}

func CheckPanicHandler(err *error) {
	if val := recover(); val != nil {
		*err = val.(error)
	}
}