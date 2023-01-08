package tictactoeaggregate

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestWining(t *testing.T) {
	useCases := map[string]struct {
		sequence []Move
		move     Move
		playerID PlayerID
		want     bool
	}{
		"winning sequence": {
			sequence: []Move{
				"1.1", "2.2", "1.2", "2.3", "1.3",
			},
			playerID: "1",
			want:     true,
		},
		"winning sequence 2": {
			sequence: []Move{
				"1.1", "2.2", "1.2", "3.1",
			},
			playerID: "1",
			want:     false,
		},
	}
	for name, uc := range useCases {
		t.Run(name, func(t *testing.T) {
			_, win := CheckIfMoveWin(
				uc.sequence,
				Wining3x3(),
			)
			assert.Equal(t, uc.want, win)
		})
	}
}

func TestGeneratedPositions(t *testing.T) {
	assert.Equal(t, Wining3x3(), GenerateWiningPositions(3, 3, 3))

	rows, cols := 5, 5
	for _, seq := range GenerateWiningPositions(3, rows, cols) {
		mm := map[Move]PlayerID{}
		for _, m := range seq {
			mm[m] = "1"
		}
		buf := strings.Builder{}
		PrintGameRC(&buf, mm, rows, cols)
		t.Log(buf.String())
	}
}

func TestMoveOrder(t *testing.T) {
	//{"SessionInGame":{"ID":"b89b2b67-a0cb-4f30-ba27-9ad80c507a71","Players":["c326c6fb-ca3b-4bb4-9842-7308c182d1cf","1bebce4c-b021-4595-95d6-9c114d789444"],"GameID":"46c66e01-2933-4b67-91b6-1903313ecfac","GameState":{"GameProgress":{"FirstPlayerID":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf","SecondPlayerID":"1bebce4c-b021-4595-95d6-9c114d789444","BoardRows":5,"BoardCols":5,"WinningLength":3,"NextMovePlayerID":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf",
	//"AvailableMoves":{"2.5":{},"3.1":{},"3.3":{},"3.4":{},"3.5":{},"4.1":{},"4.2":{},"4.3":{},"4.4":{},"4.5":{},"5.1":{},"5.2":{},"5.3":{},"5.4":{},"5.5":{}},
	//"MovesTaken":{"1.1":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf","1.2":"1bebce4c-b021-4595-95d6-9c114d789444","1.3":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf","1.4":"1bebce4c-b021-4595-95d6-9c114d789444","1.5":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf","2.1":"1bebce4c-b021-4595-95d6-9c114d789444","2.2":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf","2.3":"1bebce4c-b021-4595-95d6-9c114d789444","2.4":"c326c6fb-ca3b-4bb4-9842-7308c182d1cf","3.2":"1bebce4c-b021-4595-95d6-9c114d789444"},
	//"MovesOrder":["1.1","1.2","1.3","1.4","1.5","2.1","2.2","2.3","2.4","3.2"]}},"GameProblem":null,"PreviousGames":null}}	1673094624.0624542
	moves := []Move{
		"1.1", "1.2", "1.3", "1.4", "1.5", "2.1", "2.2", "2.3", "2.4", "3.2",
	}
	mm := ToMovesTaken(moves)

	buf := strings.Builder{}
	PrintGameRC(&buf, mm, 5, 5)
	t.Log(buf.String())

	_, result := CheckIfMoveWin(moves, GenerateWiningPositions(3, 5, 5))
	assert.Equal(t, true, result)
}
