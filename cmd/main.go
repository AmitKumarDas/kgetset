package main

import (
	testing "github.com/AmitKumarDas/kgetset/onegvkdiffschemas"
)

func main() {
	c := testing.NewTestA()
	err := c.Test()
	if err != nil {
		panic(err)
	}
}
