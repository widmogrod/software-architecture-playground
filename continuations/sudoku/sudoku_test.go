package sudoku

import (
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
	"testing"
)

func TestSudoku(t *testing.T) {
	sudoku := LoadSudoku()
	PrintSolution(sudoku)

	solver := sat.NewSolver()
	solver.AddClosures(GameConstraints(sudoku))

	result, err := solver.Solution()
	assert.NoError(t, err)

	PrintSolution(FillSolution(sudoku, result))
}
