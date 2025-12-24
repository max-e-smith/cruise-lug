/*
Copyright © 2025 Max E Smith <max.e.smith@proton.me>
*/
package main

import (
	"github.com/max-e-smith/cruise-lug/cmd"
	_ "github.com/max-e-smith/cruise-lug/cmd/get"
	_ "github.com/max-e-smith/cruise-lug/cmd/get/cruise"
)

func main() {
	cmd.Execute()
}
