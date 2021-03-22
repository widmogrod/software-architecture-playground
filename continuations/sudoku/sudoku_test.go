package sudoku

import (
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
	"testing"
)

func TestSat4(t *testing.T) {
	sudoku := LoadSudoku()
	PrintSolution(sudoku)

	vars := CreateVars(sudoku)
	//fmt.Println(vars)
	//fmt.Println("column 1", ColumnValues(sudoku, 0))
	//fmt.Println("row 2", RowsValues(sudoku, 1))

	sat := sat.NewSolver()
	sat.AddClosures(RowsUniqe(sudoku))

	sat.PrintCNF()

	result := sat.Solution()
	PrintSolution(FillSolution(sudoku, vars, result))
}
