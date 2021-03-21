package main

import (
	"fmt"
	"github.com/widmogrod/software-architecture-playground/continuations/sat"
)

func main() {
	sudoku := LoadSudoku()
	vars := CreateVars(sudoku)
	PrintSolution(sudoku)

	//fmt.Println(len(vars))

	// todo break down problem on smaller one,
	// narrow down search space!
	solve := sat.NewSolver()
	solve.AddClosures(RowsUniqe(sudoku, vars))
	solve.PrintCNF()
	solve.OptimizedSolution()
	//solve.Solution(vars...)

	//PrintSolution(FillSolution(sudoku, vars))
}

type Closure = [][]sat.Preposition

func RowsUniqe(sudoku [9][9]int, vars []*sat.BoolVar) Closure {
	var closures Closure

	index := 0
	for i := 0; i < 1; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				values := sat.Take(vars, index, 9)
				index += 9
				//break

				closures = append(closures, sat.ExactlyOne(values)...)
			}
		}
	}

	return closures
}

func CreateVars(sudoku [9][9]int) []*sat.BoolVar {
	var result []*sat.BoolVar
	for i := 0; i < 1; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				for v := 0; v < 9; v++ {
					result = append(result, sat.MkBool())
				}
			}
		}
	}
	return result
}

func FillSolution(sudoku [9][9]int, vars []*sat.BoolVar) [9][9]int {
	solution := sudoku
	index := 0
	for i := 0; i < 1; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				for v := 0; v < 9; v++ {
					value := vars[index]
					index++

					// TODO to many times it can be set
					if value.IsTrue() {
						solution[i][j] = v + 1
					}
				}
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
