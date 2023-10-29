package main

import (
	"fmt"
	"log"
)

func main() {
	b, err := BoardFromFile("problem6.txt")
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("%v\n", b)
	b.AutoSolve()
	err = b.Solved()
	//fmt.Printf("Board:\n%s\nSolved: %s\n", b, err)
	for ri := 0; ri < b.Size; ri++ {
		//ri := 1
		if b.RowPerms[ri] != nil {
			//fmt.Printf("Row %d perms:\n", ri)
			for _, pi := range *b.RowPerms[ri] {
				fmt.Printf("%v\n", b.Perms[pi])
			}
		}
	}
	fmt.Printf("Board:\n%s\nEmpty %d\nSolved: %s\n", b, b.NumEmpty, err)
	return
}

func (b *Board) PrintAllowed() {
	for ri := 0; ri < b.Size; ri++ {
		fmt.Printf("Row %d\n", ri)
		for ci := 0; ci < b.Size; ci++ {
			fmt.Printf("%d: ", ci)
			for k, _ := range b.Allowed[ri][ci] {
				fmt.Printf("%d ", k)
			}
			fmt.Printf("\n")
		}
	}
}

//TODO: inverse of naked sets
