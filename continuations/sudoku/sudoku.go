package sudoku

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
)

var maxNumbers = 9

type Board = [9][9]int

var counter = 1
var index = map[string]int{}

func Index(r, c, v int) int {
	key := fmt.Sprintf("%d.%d.%d", r, c, v)
	if idx, ok := index[key]; ok {
		return idx
	} else {
		index[key] = counter
		counter += 1
		return index[key]
	}
	//return r*3 + c*3 + v*v
}

var game = Board{
	{1, 3, 4, 0, 2, 7, 0, 8, 9},
	{0, 0, 7, 0, 0, 9, 6, 1, 4},
	{6, 0, 0, 1, 4, 8, 0, 0, 7},

	{9, 0, 2, 0, 8, 4, 0, 0, 0},
	{0, 0, 3, 0, 0, 6, 0, 0, 0},
	{0, 0, 8, 0, 5, 0, 4, 0, 0},

	{3, 9, 6, 0, 0, 0, 1, 0, 0},
	{0, 0, 0, 0, 6, 0, 0, 9, 0},
	{0, 0, 1, 0, 0, 0, 0, 0, 5},
}

//var game = Board{
//	{1, 0, 0, 0},
//	{0, 1, 0, 0},
//	{0, 0, 1, 0},
//	{0, 0, 0, 1},
//}

func LoadSudoku() Board {
	return game
}

func GameConstraints(sudoku Board) sat.Closures {
	var closures sat.Closures

	// When value is set (non-zero) then it cannot be replaced, and must be used
	for row := 0; row < cap(sudoku); row++ {
		for col := 0; col < cap(sudoku[row]); col++ {
			for v := 1; v <= maxNumbers; v++ {
				if sudoku[row][col] == v {
					closures = append(closures, []sat.Preposition{
						sat.MkLit(Index(row, col, v)),
					})
				}
			}
		}
	}

	// Each cell need to exactly-one value
	for row := 0; row < cap(sudoku); row++ {
		for col := 0; col < cap(sudoku[row]); col++ {
			var lines []*sat.BoolVar
			for v := 1; v <= maxNumbers; v++ {
				lines = append(lines, sat.MkLit(Index(row, col, v)))
			}
			closures = append(closures, sat.ExactlyOne(sat.OneOf(lines))...)
		}
	}

	// Each row needs to have exactly-one number (only one 1 in row, only one 2 in row,...)
	for row := 0; row < cap(sudoku); row++ {
		for v := 1; v <= maxNumbers; v++ {
			var lines []*sat.BoolVar
			for col := 0; col < cap(sudoku[row]); col++ {
				lines = append(lines, sat.MkLit(Index(row, col, v)))
			}
			closures = append(closures, sat.ExactlyOne(sat.OneOf(lines))...)
		}
	}

	//Each column needs to have exactly-one number (only one 1 in column, only one 2 in column,...)
	for col := 0; col < cap(sudoku); col++ {
		for v := 1; v <= maxNumbers; v++ {
			var lines []*sat.BoolVar
			for row := 0; row < cap(sudoku[col]); row++ {
				lines = append(lines, sat.MkLit(Index(row, col, v)))
			}
			closures = append(closures, sat.ExactlyOne(sat.OneOf(lines))...)
		}
	}

	return closures
}

func FillSolution(sudoku Board, solve []sat.Preposition) Board {
	solIndex := map[int]sat.Preposition{}
	for _, prep := range solve {
		if !prep.IsTrue() {
			continue
		}
		if _, found := solIndex[prep.No()]; found {
			panic("FillSolution: cannot happen!")
		}
		solIndex[prep.No()] = prep
	}

	solution := sudoku

	for row := 0; row < cap(sudoku); row++ {
		for col := 0; col < cap(sudoku[row]); col++ {
			for v := 1; v <= maxNumbers; v++ {
				if _, ok := solIndex[Index(row, col, v)]; ok {
					if solution[row][col] != 0 && solution[row][col] != v {
						panic(fmt.Sprintf(
							"FillSolution: value is fix, but trying to set value=%d in row=%d col=%d",
							v, row+1, col+1))
					}

					solution[row][col] = v
					break
				}
			}
		}
	}

	return solution
}

func PrintSolution(sudoku Board) {
	for x := 0; x < cap(sudoku); x++ {
		for y := 0; y < cap(sudoku[x]); y++ {
			fmt.Printf("%d ", sudoku[x][y])
			if y%3 == 2 {
				fmt.Printf("| ")
			}
		}

		if x%3 == 2 {
			fmt.Printf("\n----------------------- \n")
		} else {
			fmt.Println()
		}
	}
}
