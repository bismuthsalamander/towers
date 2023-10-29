package main

import (
	"fmt"
	"log"
)

// TrimPermsFromAllowed removes entries in RowPerns and ColPerms that are not
// possible because they would violate the Allowed maps. Returns true iff any
// changes were made.
func (b *Board) TrimPermsFromAllowed() bool {
	changed := false
	for ri, rp := range b.RowPerms {
		if rp == nil {
			continue
		}
		newPerms := make([]int, 0, len(*rp))
		for _, pi := range *rp {
			isPermOk := true
			for i := 0; i < b.Size; i++ {
				if !b.IsAllowed(ri, i, b.Perms[pi][i]) {
					isPermOk = false
					break
				}
			}
			if isPermOk {
				newPerms = append(newPerms, pi)
			}
		}
		if len(*rp) != len(newPerms) {
			//fmt.Printf("Replacing row %d perms - %d -> %d\n", ri, len(*rp), len(newPerms))
			b.RowPerms[ri] = &newPerms
			changed = true
		}
	}
	for ci, cp := range b.ColPerms {
		if cp == nil {
			continue
		}
		newPerms := make([]int, 0, len(*cp))
		for _, pi := range *cp {
			isPermOk := true
			for i := 0; i < b.Size; i++ {
				if !b.IsAllowed(i, ci, b.Perms[pi][i]) {
					isPermOk = false
					break
				}
			}
			if isPermOk {
				newPerms = append(newPerms, pi)
			}
		}
		if len(*cp) != len(newPerms) {
			//fmt.Printf("Replacing col %d perms - %d -> %d\n", ci, len(*cp), len(newPerms))
			b.ColPerms[ci] = &newPerms
			changed = true
		}
	}
	return changed
}

// MarkMandatory searches for cells with only one entry in Allowed and marks
// the appropriate value. Returns true iff a change was made. The redo flag
// repeats the loop if marking a mandatory cell eliminated entries in Allowed
// for other cells in the same row or column. If, for example, cell (a, b) is
// the last empty cell in its row and column, then marking (a, b) will not
// eliminate any possibilities from other cells, and repeating the loop is not
// necessary.
func (b *Board) MarkMandatory() bool {
	changed := false
	redo := false
	for ri, row := range b.Allowed {
		for ci, allowed := range row {
			if len(allowed) != 1 || b.Get(ri, ci) != EMPTY {
				continue
			}
			k := 0
			for key, _ := range allowed {
				k = key
				break
			}
			ch, nch := b.Mark(ri, ci, k)
			if ch {
				changed = true
			}
			if nch {
				redo = true
			}
		}
	}
	if redo {
		b.MarkMandatory()
	}
	return changed
}

// TrimAllowedFromPerms will eliminate a permutation from RowPerms or ColPerms
// if it is inconsistent with any cell's Allowed list. Returns true iff at
// least one permutation was eliminated.
func (b *Board) TrimAllowedFromPerms() bool {
	changed := false
	for ri := 0; ri < b.Size; ri++ {
		for ci := 0; ci < b.Size; ci++ {
			for n, _ := range b.Allowed[ri][ci] {
				//Is n allowed in slot ci in a perm for row ri?
				found := false
				if b.RowPerms[ri] != nil {
					for _, permI := range *b.RowPerms[ri] {
						if b.Perms[permI][ci] == n {
							found = true
							break
						}
					}
					if !found {
						delete(b.Allowed[ri][ci], n)
						changed = true
						continue
					}
				}
				//Is n allowed in slot ri in a perm for col ci?
				found = false
				if b.ColPerms[ci] != nil {
					for _, permI := range *b.ColPerms[ci] {
						if b.Perms[permI][ri] == n {
							found = true
							break
						}
					}
					if !found {
						delete(b.Allowed[ri][ci], n)
						changed = true
					}
				}
			}
		}
	}
	return changed
}

// AutoSolve runs all implemented solving heuristics until the puzzle is solved
// or we run out of improvements. Missing heuristics include the opposite of
// naked sets (i.e., cells X and Y are the only possible locations for numbers
// N and M, so X and Y can't have any other numbers) and pairwise permutation
// consistency between rows or columns.
func (b *Board) AutoSolve() error {
	changed := true
	for changed && b.Solved() != nil {
		fmt.Printf("New round\n")
		changed = false
		if b.MarkMandatory() {
			fmt.Printf("MM true\n")
			changed = true
		}
		if b.TrimAllowedFromPerms() {
			fmt.Printf("TAFP true\n")
			changed = true
		}
		if b.TrimPermsFromAllowed() {
			fmt.Printf("TPFA true\n")
			changed = true
		}
		if !changed {
			for n := 2; n < b.Size-1 && !changed; n++ {
				if b.TrimNakedSets(n) {
					fmt.Printf("TNS(%d) true\n", n)
					changed = true
				}
			}
		}
		if !changed {
			for n := 2; n < b.Size-1 && !changed; n++ {
				if b.TrimFoundGroups(n) {
					fmt.Printf("TFG(%d) true\n", n)
					changed = true
				}
			}
		}
	}
	return b.Solved()
}

// NumSetsEqual return strue iff the two maps contain exactly the same keys.
func NumSetsEqual(a, b map[int]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, _ := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// CheckRowNakedSet returns true iff row rowIndex contains a naked set at the
// indices specified in indices.
func (b *Board) CheckRowNakedSet(indices []int, rowIndex int) bool {
	if len(indices) == 0 {
		return false
	}
	if len(indices) != len(b.Allowed[rowIndex][indices[0]]) {
		return false
	}
	for _, idx := range indices[1:] {
		if b.Grid[rowIndex][idx] != EMPTY {
			return false
		}
		if !NumSetsEqual(b.Allowed[rowIndex][idx], b.Allowed[rowIndex][indices[0]]) {
			return false
		}
	}
	return true
}

// CheckColumnNakedSet returns true iff col colIndex contains a naked set at
// the indices specified in indices.
func (b *Board) CheckColumnNakedSet(indices []int, colIndex int) bool {
	if len(indices) == 0 {
		return false
	}
	if len(indices) != len(b.Allowed[indices[0]][colIndex]) {
		return false
	}
	for _, idx := range indices[1:] {
		if b.Grid[idx][colIndex] != EMPTY {
			return false
		}
		if !NumSetsEqual(b.Allowed[idx][colIndex], b.Allowed[indices[0]][colIndex]) {
			return false
		}
	}
	return true
}

// SliceContains returns true iff the slice haystack contains the value needle.
func SliceContains(haystack []int, needle int) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

// DisallowAll removes all entries in toRemove from the Allowed list for cell
// ri, ci. Returns true iff at least one entry was removed.
func (b *Board) DisallowAll(ri, ci int, toRemove map[int]interface{}) bool {
	changed := false
	for k, _ := range toRemove {
		if _, ok := b.Allowed[ri][ci][k]; ok {
			delete(b.Allowed[ri][ci], k)
			changed = true
		}
	}
	return changed
}

// DisallowOthers removes all numbers *not* in toKeep from the Allowed list
// for cell ri, ci. Returns true iff at least one entry was removed.
func (b *Board) DisallowOthers(ri, ci int, toKeep []int) bool {
	changed := false
	for k, _ := range b.Allowed[ri][ci] {
		canKeep := false
		for _, v := range toKeep {
			if v == k {
				canKeep = true
				break
			}
		}
		if canKeep {
			continue
		}
		if _, ok := b.Allowed[ri][ci][k]; ok {
			delete(b.Allowed[ri][ci], k)
			changed = true
		}
	}
	return changed
}

// TrimNakedSets looks at each row and column for naked sets of size n and
// makes the appropriate changes to b.Allowed if any are found. Returns true
// iff at least one change was made. A naked set (known more commonly as a
// naked pair or naked triple) occurs when, e.g., the allowed lists for cells
// A and B are [1, 2]. It allows us to eliminate 1 and 2 from the allowed lists
// of other cells in the same line.
func (b *Board) TrimNakedSets(n int) bool {
	indices := Permute(0, b.Size-1, n)
	for ri := 0; ri < b.Size; ri++ {
		for _, idxs := range indices {
			if b.CheckRowNakedSet(idxs, ri) {
				for ci := 0; ci < b.Size; ci++ {
					if SliceContains(idxs, ci) {
						continue
					}
					if b.DisallowAll(ri, ci, b.Allowed[ri][idxs[0]]) {
						return true
					}
				}
			}
		}
	}
	for ci := 0; ci < b.Size; ci++ {
		for _, idxs := range indices {
			if b.CheckColumnNakedSet(idxs, ci) {
				for ri := 0; ri < b.Size; ri++ {
					if SliceContains(idxs, ri) {
						continue
					}
					if b.DisallowAll(ri, ci, b.Allowed[idxs[0]][ci]) {
						return true
					}
				}
			}
		}
	}
	return false
}

// TrimFoundGroups looks at each row and column for found groups of size n and
// makes the appropriate changes to b.Allowed if any are found. Returns true
// iff at least one change was made. A found group (I haven't looked up the
// correct name for this heuristic) occurs when, e.g., the numbers 2 and 3 each
// have the same exact two possible homes in a given line. Since 2 and 3 must
// go in those two cells, all other numbers can be removed from their allowed
// lists. TODO: update with the correct term for "found groups!"
func (b *Board) TrimFoundGroups(n int) bool {
	changed := false
	numbers := Permute(1, b.Size, n)
	for _, nums := range numbers {
		for ri := 0; ri < b.Size; ri++ {
			if !b.CheckRowFoundGroup(nums, ri) {
				continue
			}
			for ci := 0; ci < b.Size; ci++ {
				if b.IsAllowed(ri, ci, nums[0]) {
					if b.DisallowOthers(ri, ci, nums) {
						changed = true
					}
				}
			}
		}
		for ci := 0; ci < b.Size; ci++ {
			if !b.CheckColFoundGroup(nums, ci) {
				continue
			}
			for ri := 0; ri < b.Size; ri++ {
				if b.IsAllowed(ri, ci, nums[0]) {
					if b.DisallowOthers(ri, ci, nums) {
						changed = true
					}
				}
			}
		}
	}
	return changed
}

// CheckRowFoundGroup returns true iff row rowIndex contains a found group for
// the numbers specified in numbers.
func (b *Board) CheckRowFoundGroup(numbers []int, rowIndex int) bool {
	numberCells := make([]map[int]interface{}, len(numbers))
	for i, _ := range numbers {
		numberCells[i] = make(map[int]interface{})
	}
	for coli := 0; coli < b.Size; coli++ {
		for nidx, num := range numbers {
			if b.IsAllowed(rowIndex, coli, num) {
				numberCells[nidx][coli] = nil
			}
		}
	}
	if len(numberCells[0]) != len(numbers) {
		return false
	}
	for i := 1; i < len(numbers); i++ {
		if !NumSetsEqual(numberCells[i], numberCells[0]) {
			return false
		}
	}
	return true
}

// CheckColFoundGroup returns true iff col colIndex contains a found group for
// the numbers specified in numbers.
func (b *Board) CheckColFoundGroup(numbers []int, colIndex int) bool {
	numberCells := make([]map[int]interface{}, len(numbers))
	for i, _ := range numbers {
		numberCells[i] = make(map[int]interface{})
	}
	for rowi := 0; rowi < b.Size; rowi++ {
		for nidx, num := range numbers {
			if b.IsAllowed(rowi, colIndex, num) {
				numberCells[nidx][rowi] = nil
			}
		}
	}
	if len(numberCells[0]) != len(numbers) {
		return false
	}
	for i := 1; i < len(numbers); i++ {
		if !NumSetsEqual(numberCells[i], numberCells[0]) {
			return false
		}
	}
	return true
}

func testRowFoundGroup() {
	str := "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	b, err := BoardFromString(str)
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("sz %d\n%s\n", b.Size, b)
	b.DisallowOthers(0, 0, []int{2, 3, 4})
	b.DisallowOthers(0, 1, []int{2, 3, 5})
	b.DisallowOthers(0, 2, []int{1, 4, 5})
	b.DisallowOthers(0, 3, []int{1, 4, 5})
	b.DisallowOthers(0, 4, []int{1, 4, 5})
	b.PrintAllowed()
	res := b.TrimFoundGroups(2)
	fmt.Printf("%v\n", res)
	fmt.Printf("sz %d\n%s\n", b.Size, b)
	b.PrintAllowed()
}

func testColFoundGroup() {
	str := "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	str += "       \n"
	b, err := BoardFromString(str)
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("sz %d\n%s\n", b.Size, b)
	b.DisallowOthers(0, 0, []int{2, 3, 4})
	b.DisallowOthers(1, 0, []int{2, 3, 5})
	b.DisallowOthers(2, 0, []int{1, 4, 5})
	b.DisallowOthers(3, 0, []int{1, 4, 5})
	b.DisallowOthers(4, 0, []int{1, 4, 5})
	b.PrintAllowed()
	res := b.TrimFoundGroups(2)
	fmt.Printf("%v\n", res)
	fmt.Printf("sz %d\n%s\n", b.Size, b)
	b.PrintAllowed()
}
