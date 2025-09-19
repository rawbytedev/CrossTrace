package main

import (
	"crosstrace/internal/configs"
	"fmt"
)

func main() {
	fmt.Print("Generating Configs")
	configs.GeneDefault()
}
