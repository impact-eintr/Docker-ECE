package common

func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
