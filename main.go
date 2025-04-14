package main

import "github.com/hambosto/go-encryption/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
