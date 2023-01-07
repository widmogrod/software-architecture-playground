export function CreateGameCMD(p1, wh, len) {
    return {
        "CreateGameCMD": {
            "FirstPlayerID": p1,
            "BoardRows": wh,
            "BoardCols": wh,
            "WinningLength": len,
        }
    }
}

export function JoinGameCMD(p2) {
    return {
        "StartGameCMD": {
            "SecondPlayerID": p2,
        }
    }
}

export function StartGameCMD(p1, p2, wh, len) {
    return {
        "StartGameCMD": {
            "FirstPlayerID": p1,
            "SecondPlayerID": p2,
            "BoardRows": wh | 0,
            "BoardCols": wh | 0,
            "WinningLength": len | 0,
        }
    }
}

export function MoveCMD(pid, position) {
    return {
        "MoveCMD": {
            "PlayerID": pid,
            "Position": position,
        }
    }
}