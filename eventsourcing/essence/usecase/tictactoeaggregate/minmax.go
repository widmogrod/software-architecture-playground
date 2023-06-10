package tictactoeaggregate

import (
	"fmt"
	"math"
	"sort"
)

func NextMoveNaive(taken map[Move]PlayerID, rows, cols, len int) Move {
	for r := 1; r <= rows; r++ {
		for c := 1; c <= cols; c++ {
			m := MkMove(r, c)
			if _, ok := taken[m]; !ok {
				return m
			}
		}
	}

	return ""
}

//func SlidingBoards(moves []Move, rows, cols int) ([][][]PlayerID, PlayerID) {
//
//	masterBoard := make([][]PlayerID, rows)
//	for r := 1; r <= rows; r++ {
//		masterBoard[r-1] = make([]PlayerID, cols)
//		for c := 1; c <= cols; c++ {
//			masterBoard[r-1][c-1] = ""
//		}
//	}
//
//	var lastPlayer = PlayerID("O")
//
//	for i, m := range moves {
//		var r, c int
//		_, err := fmt.Sscanf(m, "%d.%d", &r, &c)
//		if err != nil {
//			panic(err)
//		}
//
//		if i%2 == 0 {
//			masterBoard[r-1][c-1] = "O"
//		} else {
//			masterBoard[r-1][c-1] = "X"
//		}
//
//		lastPlayer = masterBoard[r-1][c-1]
//	}
//
//	var boards [][][]PlayerID
//	len := 3
//	for shiftR := 0; shiftR < rows-len; shiftR++ {
//		for shiftC := 0; shiftC < cols-len; shiftC++ {
//			board := make([][]PlayerID, len)
//			for r := shiftR; r <= shiftR+len; r++ {
//				row := make([]PlayerID, len)
//				for c := shiftC; c <= shiftC+len; c++ {
//					row = append(row, masterBoard[r][c])
//				}
//				board = append(board, row)
//			}
//
//			boards = append(boards, board)
//		}
//	}
//
//	return boards, lastPlayer
//}
//
//func SlidingNextMoveMinMax(moves []Move, rows, cols int) Move {
//	boards, lastPlayer := SlidingBoards(moves, rows, cols)
//
//	var (
//		bestMove  Move
//		bestScore float64 = -1000
//	)
//
//	for _, board := range boards {
//		for r := 1; r <= rows; r++ {
//			for c := 1; c <= cols; c++ {
//				m := MkMove(r, c)
//				if board[r-1][c-1] == "" {
//					board[r-1][c-1] = nextPlayer(lastPlayer)
//					score := minimax(board, nextPlayer(lastPlayer), true, 0)
//					board[r-1][c-1] = ""
//					if score > bestScore {
//						bestScore = score
//						bestMove = m
//					}
//				}
//			}
//		}
//	}
//
//	return bestMove
//}

func NextMoveMinMax(moves []Move, rows, cols int) Move {
	var (
		bestMove   Move
		bestScore  float64 = -1000
		lastPlayer         = PlayerID("O")
	)

	board := make([][]PlayerID, rows)
	for r := 1; r <= rows; r++ {
		board[r-1] = make([]PlayerID, cols)
		for c := 1; c <= cols; c++ {
			board[r-1][c-1] = ""
		}
	}

	for i, m := range moves {
		var r, c int
		_, err := fmt.Sscanf(m, "%d.%d", &r, &c)
		if err != nil {
			panic(err)
		}

		if i%2 == 0 {
			board[r-1][c-1] = "O"
		} else {
			board[r-1][c-1] = "X"
		}

		lastPlayer = board[r-1][c-1]
	}

	for r := 1; r <= rows; r++ {
		for c := 1; c <= cols; c++ {
			m := MkMove(r, c)
			if board[r-1][c-1] == "" {
				board[r-1][c-1] = nextPlayer(lastPlayer)
				score := minimax(board, nextPlayer(lastPlayer), true, 6, -1000, 1000)
				board[r-1][c-1] = ""
				if score > bestScore {
					bestScore = score
					bestMove = m
				}
			}
		}
	}

	return bestMove
}

func minimax(board [][]PlayerID, player PlayerID, isMaximising bool, depth int, a, b float64) float64 {
	if depth == 0 {
		return 0
	}

	if won, winner := isWin(board); won {
		var score float64
		if winner == player {
			score = 1
		} else {
			score = -1
		}

		if !isMaximising {
			score = -score
		}

		return score * float64(depth)
	} else if isTie(board) {
		return 0
	}

	if !isMaximising {
		var maxEval float64 = -1000
		for r := 1; r <= len(board); r++ {
			for c := 1; c <= len(board[r-1]); c++ {
				if board[r-1][c-1] == "" {
					board[r-1][c-1] = nextPlayer(player)
					eval := minimax(board, nextPlayer(player), !isMaximising, depth-1, a, b)
					board[r-1][c-1] = ""

					maxEval = math.Max(maxEval, eval)
					if maxEval >= b {
						break
					}

					a = math.Max(a, eval)
				}
			}
		}
		return maxEval
	} else {
		var minEval float64 = 1000
		for r := 1; r <= len(board); r++ {
			for c := 1; c <= len(board[r-1]); c++ {
				if board[r-1][c-1] == "" {
					board[r-1][c-1] = nextPlayer(player)
					eval := minimax(board, nextPlayer(player), !isMaximising, depth-1, a, b)
					board[r-1][c-1] = ""

					minEval = math.Min(minEval, eval)
					if minEval <= a {
						break
					}

					b = math.Min(b, eval)
				}
			}
		}
		return minEval
	}
}

func isTie(board [][]PlayerID) bool {
	for r := 1; r <= len(board); r++ {
		for c := 1; c <= len(board[r-1]); c++ {
			if board[r-1][c-1] == "" {
				return false
			}
		}
	}
	return true
}

var boardIsWonResult = map[string]bool{}
var boardIsWonPlayer = map[string]PlayerID{}

func serialiseBoard(board [][]PlayerID) string {
	var s string
	for r := 1; r <= len(board); r++ {
		for c := 1; c <= len(board[r-1]); c++ {
			s += string(board[r-1][c-1])
			s += ","
		}
	}
	return s
}

func cachedIsWin(board [][]PlayerID) (bool, PlayerID) {
	s := serialiseBoard(board)
	if _, ok := boardIsWonResult[s]; ok {
		return boardIsWonResult[s], boardIsWonPlayer[s]
	}

	won, winner := isWin(board)
	boardIsWonResult[s] = won
	boardIsWonPlayer[s] = winner
	return won, winner
}

func isWin(board [][]PlayerID) (won bool, winner PlayerID) {
	var movesO []Move
	var movesX []Move

	for r := 1; r <= len(board); r++ {
		for c := 1; c <= len(board[r-1]); c++ {
			if board[r-1][c-1] == "O" {
				movesO = append(movesO, MkMove(r, c))
			} else if board[r-1][c-1] == "X" {
				movesX = append(movesX, MkMove(r, c))
			}
		}
	}

	if len(movesO) < 3 && len(movesX) < 3 {
		return false, ""
	}

	sort.Strings(movesO)
	sort.Strings(movesX)

	winseq := CachedGenerateWiningPositions(3, len(board), len(board[0]))
	for _, seq := range winseq {
		if _, won := is_win_seq(movesO, seq); won {
			return true, "O"
		}
		if _, won := is_win_seq(movesX, seq); won {
			return true, "X"
		}
	}

	return false, ""
}

func nextPlayer(p string) string {
	if p == "X" {
		return "O"
	}
	return "X"
}
