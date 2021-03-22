package sudoku

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
)

var maxNumbers = 3

func Index(r, c, v int) int {
	return r*1 + c*3 + v
}

func ForEmpty(sudoku Board, f func(int, int, int)) {
	for x := 0; x < cap(sudoku); x++ {
		for y := 0; y < cap(sudoku[x]); y++ {
			for v := 0; v < maxNumbers; v++ {
				f(x, y, v)
			}
		}
	}
}

func RowsValues(sudoku Board, row int) map[int]int {
	result := map[int]int{}
	for _, v := range sudoku[row] {
		if v != 0 {
			result[v] = v
		}
	}
	return result
}

func ColumnValues(sudoku Board, column int) map[int]int {
	result := map[int]int{}
	for x := 0; x < cap(sudoku); x++ {
		if sudoku[x][column] != 0 {
			v := sudoku[x][column]
			result[v] = v
		}
	}
	return result
}

func RowsUniqe(sudoku Board) sat.Closures {
	var closures sat.Closures

	for x := 0; x < cap(sudoku); x++ {
		for v := 0; v < maxNumbers; v++ {
			var lines []*sat.BoolVar
			for y := 0; y < cap(sudoku[x]); y++ {
				lines = append(lines, sat.MkLit(Index(x, y, v)))
			}
			closures = append(closures, sat.ExactlyOne(lines)...)
		}
	}

	//for x := 0; x < cap(sudoku); x++ {
	//	for y := 0; y < cap(sudoku[x]); y++ {
	//		var lines []*sat.BoolVar
	//		for v := 0; v < maxNumbers; v++ {
	//			lines = append(lines, sat.MkLit(Index(x, y, v)))
	//		}
	//		closures = append(closures, sat.ExactlyOne(lines)...)
	//	}
	//}

	return closures
}

func CreateVars(sudoku Board) []*sat.BoolVar {
	var result []*sat.BoolVar
	ForEmpty(sudoku, func(x int, y int, v int) {
		result = append(result, sat.MkLit(Index(x, y, v)))
	})
	return result
}

func FillSolution(sudoku Board, vars []*sat.BoolVar, solve []sat.Preposition) Board {
	solIndex := map[int]sat.Preposition{}
	for _, prep := range solve {
		if _, found := solIndex[prep.No()]; found {
			continue
			//panic("cannot happen!")
		}
		solIndex[prep.No()] = prep
	}

	solution := sudoku
	ForEmpty(sudoku, func(x int, y int, v int) {
		if _, ok := solIndex[Index(x, y, v)]; ok {
			solution[x][y] = v + 1
			return
		}
	})

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

type Board = [1][3]int

var game = Board{
	{1, 0, 3},
	//{3, 0, 4},
	//{0, 6, 0},
}

func LoadSudoku() Board {
	return game
}
