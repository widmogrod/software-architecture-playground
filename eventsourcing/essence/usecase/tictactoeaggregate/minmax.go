package tictactoeaggregate

import (
	"fmt"
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

	//if lastPlayer != "O" {
	//	panic("last player must be O")
	//}

	for r := 1; r <= rows; r++ {
		for c := 1; c <= cols; c++ {
			m := MkMove(r, c)
			if board[r-1][c-1] == "" {
				board[r-1][c-1] = nextPlayer(lastPlayer)
				score := minmax(board, nextPlayer(lastPlayer), true, 0)
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

func minmax(board [][]PlayerID, player PlayerID, isMaximising bool, depth float64) float64 {
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

		return score
	} else if isTie(board) {
		return 0
	}

	if depth > 4 {
		return 0
	}

	if !isMaximising {
		var bestScore float64 = -1000
		for r := 1; r <= len(board); r++ {
			for c := 1; c <= len(board[r-1]); c++ {
				if board[r-1][c-1] == "" {
					board[r-1][c-1] = nextPlayer(player)
					score := minmax(board, nextPlayer(player), true, depth+1)
					board[r-1][c-1] = ""
					if score > bestScore {
						bestScore = score
					}
				}
			}
		}
		return bestScore
	} else {
		var bestScore float64 = 1000
		for r := 1; r <= len(board); r++ {
			for c := 1; c <= len(board[r-1]); c++ {
				if board[r-1][c-1] == "" {
					board[r-1][c-1] = nextPlayer(player)
					score := minmax(board, nextPlayer(player), false, depth+1)
					board[r-1][c-1] = ""
					if score < bestScore {
						bestScore = score
					}
				}
			}
		}
		return bestScore
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
