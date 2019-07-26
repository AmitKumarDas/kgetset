package main

import (
  k8s "github.com/AmitKumarDas/kgetset"
)

func main() {
	c := k8s.HelloTestA()
	err := c.Test()
	if err != nil {
		panic(err)
	}
}
