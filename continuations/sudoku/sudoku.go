package sudoku

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
)

var maxNumbers = 3

func ForEmpty(sudoku Board, f func(int, int, int, int)) {
	index := 0
	for x := 0; x < cap(sudoku); x++ {
		for y := 0; y < cap(sudoku[x]); y++ {
			if sudoku[x][y] == 0 {
				for v := 0; v < maxNumbers; v++ {
					f(x, y, index, v)
					index += 1
				}
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

func ColumnUniqe(sudoku Board, vars []*sat.BoolVar) sat.Closures {
	var closures sat.Closures

	ForEmpty(sudoku, func(x int, y int, index, v int) {
		if v == 0 {
			values := sat.Take(vars, index, maxNumbers)
			closures = append(closures, sat.ExactlyOne(values)...)
		}
	})

	return closures
}

func columnUnique() {

}

func CreateVars(sudoku Board) []*sat.BoolVar {
	var result []*sat.BoolVar
	ForEmpty(sudoku, func(x int, y int, index, v int) {
		result = append(result, sat.MkBoolC(index))
	})
	return result
}

func FillSolution(sudoku Board, vars []*sat.BoolVar, solve []sat.Preposition) Board {
	solIndex := map[int]sat.Preposition{}
	for _, prep := range solve {
		if _, found := solIndex[prep.No()]; found {
			panic("cannot happen!")
		}
		solIndex[prep.No()] = prep
	}

	solution := sudoku
	ForEmpty(sudoku, func(x int, y int, index, v int) {
		if _, ok := solIndex[index]; ok {
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

type Board = [3][3]int

var game = Board{
	{1, 2, 0},
	{3, 0, 4},
	{0, 6, 0},
}

func LoadSudoku() Board {
	return game
}
