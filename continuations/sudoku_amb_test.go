package continuations

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/amb"
	"testing"
)

func TestSudokuAmbTest(t *testing.T) {
	sudoku := LoadSudoku()
	vars := CreateVars(sudoku)

	solve := amb.NewRuntime()
	solve.With(vars...)
	solve.Until(func() bool {
		_ = FillSolution(sudoku, vars)
		//return RowsUniqe(solution)
		return true
	})

	PrintSolution(FillSolution(sudoku, vars))
}

func RowsUniqe(sudoku [9][9]int) bool {
	for i := 0; i < 9; i++ {
		unique := map[int]struct{}{}
		for j := 0; j < 9; j++ {
			value := sudoku[i][j]
			if _, found := unique[value]; found {
				return false
			}
			unique[value] = struct{}{}
		}
	}

	return true
}

func CreateVars(sudoku [9][9]int) []*amb.Value {
	var result []*amb.Value
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				result = append(result, amb.MkRange(1, 9))
			}
		}
	}
	return result
}

func FillSolution(sudoku [9][9]int, vars []*amb.Value) [9][9]int {
	solution := sudoku
	index := 0
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				value := vars[index]
				index++
				solution[i][j] = value.Val()
			}
		}
	}
	return solution
}

func PrintSolution(sudoku [9][9]int) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Printf("%d ", sudoku[i][j])
			if j%3 == 2 {
				fmt.Printf("| ")
			}
		}
		fmt.Printf("\n")
		if i%3 == 2 {
			fmt.Printf("----------------------- \n")
		}
	}
}

var game = [9][9]int{
	{5, 3, 0, 0, 7, 0, 0, 0, 0},
	{6, 0, 0, 1, 9, 5, 0, 0, 0},
	{0, 9, 8, 0, 0, 0, 0, 6, 0},
	{8, 0, 0, 0, 6, 0, 0, 0, 3},
	{4, 0, 0, 8, 0, 3, 0, 0, 1},
	{7, 0, 0, 0, 2, 0, 0, 0, 6},
	{0, 6, 0, 0, 0, 0, 2, 8, 0},
	{0, 0, 0, 4, 1, 9, 0, 0, 5},
	{0, 0, 0, 0, 8, 0, 0, 7, 9},
}

func LoadSudoku() [9][9]int {
	return game
}
