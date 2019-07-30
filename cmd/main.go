package main

import (
  "github.com/AmitKumarDas/kgetset/hello"
)

func main() {
	c := hello.NewTestA()
	err := c.Test()
	if err != nil {
		panic(err)
	}
}
