package kgetset

func main() {
	c := newCRD()
	err := c.test()
	if err != nil {
		panic(err)
	}
}
