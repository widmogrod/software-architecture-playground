package sudoku

import (
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
	"testing"
)

func TestSat4(t *testing.T) {
	sudoku := LoadSudoku()
	PrintSolution(sudoku)

	solver := sat.NewSolver()
	solver.AddClosures(GameConstraints(sudoku))
	solver.PrintCNF()
	result := solver.Solution()

	PrintSolution(FillSolution(sudoku, result))
}
