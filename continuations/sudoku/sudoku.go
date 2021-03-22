package sudoku

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
)

var maxNumbers = 3

func Index(r, c, v int) int {
	return r*3 + c*3 + v
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

	// only one value in row can be true
	for row := 0; row < cap(sudoku); row++ {
		for v := 1; v <= maxNumbers; v++ {
			var lines []*sat.BoolVar
			for col := 0; col < cap(sudoku[row]); col++ {
				lines = append(lines, sat.MkLit(Index(row, col, v)))
			}
			closures = append(closures, sat.ExactlyOne(lines)...)
		}
	}

	// only one value per column can be true
	for row := 0; row < cap(sudoku); row++ {
		for col := 0; col < cap(sudoku[row]); col++ {
			var lines []*sat.BoolVar
			for v := 1; v <= maxNumbers; v++ {
				lines = append(lines, sat.MkLit(Index(row, col, v)))
			}
			closures = append(closures, sat.ExactlyOne(lines)...)
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
			panic("cannot happen!")
		}
		solIndex[prep.No()] = prep
	}

	solution := sudoku

	for row := 0; row < cap(sudoku); row++ {
		for col := 0; col < cap(sudoku[row]); col++ {
			for v := 1; v <= maxNumbers; v++ {
				if _, ok := solIndex[Index(row, col, v)]; ok {
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

type Board = [3][3]int

var game = Board{
	{1, 0, 3},
	{2, 0, 1},
	{0, 1, 0},
}

func LoadSudoku() Board {
	return game
}
