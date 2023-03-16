package main

import (
	"fmt"
	"os/exec"
)


func mai(){
	cmd := exec.Command("go", "vet", "./...")
	out, err :=  cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			print(string(out))
		}
	}
	fmt.Println(string(out))

}