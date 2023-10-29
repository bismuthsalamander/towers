package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	OBS_ROW int = 0
	OBS_COL int = 1
	OBS_FWD int = 0
	OBS_BWD int = 1
	EMPTY   int = 0
)

// An Observer embodies a row or column constraint. Type is either OBS_ROW or
// OBS_COL, and Direction is either OBS_FWD for increasing indices (i.e.,
// the observer is looking from left to right or top to bottom) and OBS_BWD
// for decreasing indices (right to left or bottom to top).
type Observer struct {
	Type      int
	Index     int
	Direction int
	Count     int
}

// A Board stores the current state of the problem, including all structures
// used to track partial solution data. ObsSorted contains pointers to Observer
// objects in the following order:
//
//   - Row 0's OBS_FWD observer
//   - Row 0's OBS_BWD observer
//   - Row 1's OBS_FWD observer
//   - ...
//   - Row n's OBS_FWD observer
//   - Row n's OBS_BWD observer
//   - Col 0's OBS_FWD observer
//   - Col 0's OBS_BWD observer
//   - Col 1's OBS_FWD observer
//
// - ...
//   - Col n's OBS_BWD observer
//
// Perms contains a slice of slices representing all permutations of the
// numbers 1 to BoardSize, inclusive. RowPerms and ColPerms contain, for each
// row or column, a slice of indices into Perms representing the permutations
// that are possible for that row or column.
type Board struct {
	Grid      [][]int
	Allowed   [][]map[int]interface{}
	NumEmpty  int
	Size      int
	Observers []*Observer
	ObsSorted []*Observer
	Perms     [][]int
	RowPerms  []*[]int
	ColPerms  []*[]int
}

// PermsForObs generates a slice of the permutation indexes that fit both
// observers. If both are nil, returns nil. Must be called after b.Perms has
// been initialized.
func (b *Board) PermsForObs(fwd, bwd *Observer) *[]int {
	if fwd == nil && bwd == nil {
		return nil
	}
	out := make([]int, 0)
	for i := 0; i < len(b.Perms); i++ {
		if PermFitsObs(b.Perms[i], fwd, bwd) {
			out = append(out, i)
		}
	}
	return &out
}

// PermFitsObs checks whether a given row or column is consistent with both
// observers. Nil inputs are ignored, so PermFitsObs(_, nil, nil) always
// returns true.
func PermFitsObs(p []int, fwd, bwd *Observer) bool {
	if fwd != nil {
		vis := 0
		highest := 0
		for i := 0; i < len(p); i++ {
			if p[i] > highest {
				highest = p[i]
				vis++
				if vis > fwd.Count {
					return false
				}
			}
		}
		if vis != fwd.Count {
			return false
		}
	}
	if bwd != nil {
		vis := 0
		highest := 0
		for i := len(p) - 1; i >= 0; i-- {
			if p[i] > highest {
				highest = p[i]
				vis++
				if vis > bwd.Count {
					return false
				}
			}
		}
		if vis != bwd.Count {
			return false
		}
	}
	return true
}

// PopulateRowColPerms is used during initialization to generate the lists of
// allowed permutations for each row and column.
func (b *Board) PopulateRowColPerms() {
	pi := 0
	for ri := 0; ri < b.Size; ri++ {
		b.RowPerms[ri] = b.PermsForObs(b.ObsSorted[pi], b.ObsSorted[pi+1])
		pi += 2
	}
	for ci := 0; ci < b.Size; ci++ {
		b.ColPerms[ci] = b.PermsForObs(b.ObsSorted[pi], b.ObsSorted[pi+1])
		pi += 2
	}
}

// Get returns the grid value at the specified coordinates.
func (b *Board) Get(ri, ci int) int {
	return b.Grid[ri][ci]
}

// ObserverSatisfied returns false if the grid's contents are consistent with
// the constraint for the specified observer. This function will treat empty
// cells as a zero (meaning that such cells are never visible and never
// obstruct other cells), so the return value may be misleading if called when
// the relevant row or column is incomplete.
func (b *Board) ObserverSatisfied(o *Observer) bool {
	ri := 0
	ci := 0
	dr := 0
	dc := 0
	if o.Type == OBS_ROW {
		ri = o.Index
		dc = 1
		if o.Direction == OBS_BWD {
			ci = b.Size - 1
			dc = -1
		}
	} else if o.Type == OBS_COL {
		ci = o.Index
		dr = 1
		if o.Direction == OBS_BWD {
			ri = b.Size - 1
			dr = -1
		}
	}
	vis := 0
	highest := 0
	for i := 0; i < b.Size; i++ {
		val := b.Get(ri, ci)
		if val > highest {
			vis++
			highest = val
		}
		ri += dr
		ci += dc
	}
	return vis == o.Count
}

// AddObserver seeds the Observer object into Observers and into ObsSorted at
// the correct index.
func (b *Board) AddObserver(o *Observer) {
	if o.Count == 0 {
		return
	}
	b.Observers = append(b.Observers, o)
	ind := 0
	if o.Type == OBS_ROW {
		ind = o.Index * 2
	} else {
		ind = b.Size*2 + (o.Index * 2)
	}
	if o.Direction == OBS_BWD {
		ind += 1
	}
	if b.ObsSorted[ind] != nil {
		panic(fmt.Sprintf("AddObserver is replacing observer %s with new observer %s", b.ObsSorted[ind], o))
	}
	b.ObsSorted[ind] = o
}

func (o Observer) String() string {
	out := ""
	if o.Type == OBS_ROW {
		out += fmt.Sprintf("row %d", o.Index)
	} else {
		out += fmt.Sprintf("col %d", o.Index)
	}
	if o.Direction == OBS_BWD {
		out += " BWD"
	}
	out += fmt.Sprintf(" sees %d", o.Count)
	return out
}

// Solved returns true iff all observers are satisfied and all cells are
// filled. As of now, it does not confirm that the sudoku rule is satisfied.
func (b *Board) Solved() error {
	if b.NumEmpty != 0 {
		return fmt.Errorf("grid has %d empty cells; need 0", b.NumEmpty)
	}
	for _, o := range b.Observers {
		if !b.ObserverSatisfied(o) {
			return fmt.Errorf("observer %s unsatisfied", o)
		}
	}
	return nil
}

// Mark sets cell at row ri, col ci as val. Return values are:
//   - true iff the cell was changed
//   - true iff a neighbor of the updated cell had val removed from its
//     allowed list
func (b *Board) Mark(ri, ci, val int) (bool, bool) {
	neighborUpdated := false
	if !b.Set(ri, ci, val) {
		return false, false
	}
	for i := 0; i < b.Size; i++ {
		if i != ri && b.IsAllowed(i, ci, val) {
			delete(b.Allowed[i][ci], val)
			neighborUpdated = true
		}
		if i != ci && b.IsAllowed(ri, i, val) {
			delete(b.Allowed[ri][i], val)
			neighborUpdated = true
		}
	}
	for i := 1; i <= b.Size; i++ {
		if i != val {
			delete(b.Allowed[ri][ci], i)
		}
	}
	return true, neighborUpdated
}

// Unset is a shortcut for Set(ri, ci, EMPTY).
func (b *Board) Unset(ri, ci int) bool {
	return b.Set(ri, ci, EMPTY)
}

// IsAllowed queries the Allowed list for the specified cell, returning a bool.
func (b *Board) IsAllowed(ri, ci, n int) bool {
	_, ok := b.Allowed[ri][ci][n]
	return ok
}

// IntToCh generates a rune representing a number, starting with digits and
// continuing through lowercase letters.
func IntToCh(n int) rune {
	if n > 9 {
		return 'a' + rune(n-10)
	}
	return '0' + rune(n)
}

// ChToInt reverses IntToCh, parsing a rune and turning it into an int.
func ChToInt(ch rune) int {
	if ch >= '1' && ch <= '9' {
		return int(ch - '0')
	} else if ch >= 'a' && ch <= 'z' {
		return int(ch-'a') + 10
	}
	return 0
}

// BoardFromFile takes a filename as input and generates a board from it.
func BoardFromFile(f string) (*Board, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return BoardFromString(string(data))
}

// BoardFromString takes an input string and parses it into a board.
func BoardFromString(input string) (*Board, error) {
	b := Board{}
	lines := make([]string, 0)
	inputs := make([][]int, 0)
	for _, txt := range strings.Split(input, "\n") {
		txt = strings.Trim(txt, "\r\n")
		if len(txt) > 0 {
			lines = append(lines, txt)
		}
	}
	b.Size = len(lines) - 2
	b.Allowed = NewAllowed(b.Size)
	b.NumEmpty = b.Size * b.Size
	b.Observers = make([]*Observer, 0, b.Size*4)
	b.ObsSorted = make([]*Observer, b.Size*4)
	b.RowPerms = make([]*[]int, b.Size)
	b.ColPerms = make([]*[]int, b.Size)
	b.Grid = make([][]int, b.Size)
	for i := 0; i < b.Size; i++ {
		b.Grid[i] = make([]int, b.Size)
	}
	for i := 0; i < b.Size+2; i++ {
		inputs = append(inputs, make([]int, b.Size+2))
	}
	for ri, row := range lines {
		for ci, ch := range row {
			inputs[ri][ci] = ChToInt(ch)
		}
	}

	for ri, row := range inputs {
		for ci, cell := range row {
			if ri == 0 || ri == b.Size+1 {
				if ci == 0 || ci == b.Size+1 {
					continue
				}
				obs := Observer{
					Type:      OBS_COL,
					Index:     ci - 1,
					Direction: OBS_FWD,
					Count:     cell,
				}
				if ri == b.Size+1 {
					obs.Direction = OBS_BWD
				}
				b.AddObserver(&obs)
				continue
			}
			if ci == 0 || ci == b.Size+1 {
				obs := Observer{
					Type:      OBS_ROW,
					Index:     ri - 1,
					Direction: OBS_FWD,
					Count:     cell,
				}
				if ci == b.Size+1 {
					obs.Direction = OBS_BWD
				}
				b.AddObserver(&obs)
				continue
			}
			b.Mark(ri-1, ci-1, cell)
		}
	}
	b.Perms = PermuteN(b.Size)
	b.PopulateRowColPerms()
	b.TrimAllowedFromPerms()
	fmt.Printf("After init, numEmpty %d\n", b.NumEmpty)
	return &b, nil
}

// ObsChar is a helper function that locates the observer specified by the
// t(ype), index and direction parameters, then returns a string to be
// displayed in the board string.
func (b *Board) ObsChar(t, index, direction int) string {
	idx := index * 2
	if t == OBS_COL {
		idx += b.Size * 2
	}
	if direction == OBS_BWD {
		idx += 1
	}
	o := b.ObsSorted[idx]
	if o == nil {
		return " "
	}
	return string(IntToCh(o.Count))
}

// CharAt generates a character for the specified cell in the board's grid.
func (b *Board) CharAt(ri, ci int) string {
	if ri < 0 || ci < 0 || ri >= b.Size || ci >= b.Size {
		return " "
	}
	return string(IntToCh(b.Get(ri, ci)))
}

func (b *Board) String() string {
	out := " "
	for ci := 0; ci < b.Size; ci++ {
		out += b.ObsChar(OBS_COL, ci, OBS_FWD)
	}
	out += "\n"
	for ri := 0; ri < b.Size; ri++ {
		out += b.ObsChar(OBS_ROW, ri, OBS_FWD)
		for ci := 0; ci < b.Size; ci++ {
			out += b.CharAt(ri, ci)
		}
		out += b.ObsChar(OBS_ROW, ri, OBS_BWD)
		out += "\n"
	}
	out += " "
	for ci := 0; ci < b.Size; ci++ {
		out += b.ObsChar(OBS_COL, ci, OBS_BWD)
	}
	return out
}

func (b *Board) PrintGrid() {
	for _, row := range b.Grid {
		for _, cell := range row {
			fmt.Printf("%d", cell)
		}
		fmt.Printf("\n")
	}
}

// Set saves an entry in the grid and updates NumEmpty.
func (b *Board) Set(ri, ci, val int) bool {
	if b.Get(ri, ci) == val {
		return false
	}
	if b.Get(ri, ci) != EMPTY && val == EMPTY {
		b.NumEmpty++
	} else if b.Get(ri, ci) == EMPTY && val != EMPTY {
		b.NumEmpty--
	}
	b.Grid[ri][ci] = val
	return true
}

// NumSet generates a set (i.e., a map[int]interface{}) containing the positive
// integers from 1 to n inclusive.
func NumSet(n int) map[int]interface{} {
	out := make(map[int]interface{})
	for i := 1; i <= n; i++ {
		out[i] = nil
	}
	return out
}

// NewAllowed populates the Allowed slice with a new set from NumSet for each
// location.
func NewAllowed(n int) [][]map[int]interface{} {
	out := make([][]map[int]interface{}, 0)
	for ri := 0; ri < n; ri++ {
		out = append(out, make([]map[int]interface{}, 0))
		for ci := 0; ci < n; ci++ {
			out[ri] = append(out[ri], NumSet(n))
		}
	}
	return out
}
