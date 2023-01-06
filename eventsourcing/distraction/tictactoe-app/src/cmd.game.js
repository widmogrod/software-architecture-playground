export function CreateGameCMD(p1) {
    return {
        "CreateGameCMD": {
            "FirstPlayerID": p1,
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

export function StartGameCMD(p1, p2) {
    return {
        "StartGameCMD": {
            "FirstPlayerID": p1,
            "SecondPlayerID": p2,
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