export function CreateSessionCMD(sid, players) {
    return {
        "CreateSessionCMD": {
            "SessionID": sid,
            "NeedsPlayers": players || 2,
        },
    }
}

export function JoinGameSessionCMD(sid, pid) {
    return {
        "JoinGameSessionCMD": {
            "SessionID": sid,
            "PlayerID": pid,
        },
    }
}

export function GameSessionWithBotCMD(sid) {
    return {
        "GameSessionWithBotCMD": {
            "SessionID": sid,
        },
    }
}

export function NewGameCMD(sid, gid) {
    return {
        "NewGameCMD": {
            "SessionID": sid,
            "GameID": gid,
        },
    }
}

export function GameActionCMD(sid, gid, action) {
    return {
        "GameActionCMD": {
            "SessionID": sid,
            "GameID": gid,
            "Action": action,
            // Action: JSON.stringify(action),
        },
    }
}

export function SequenceCMD(cmds) {
    return {
        "SequenceCMD": {
            "Commands": cmds || [],
        },
    }
}

export function SessionStatsQuery(sessionID) {
    return {
        "SessionStatsQuery": {
            "SessionID": sessionID,
        },
    }
}

export function SessionStatsSubscription(sessionID) {
    return {
        "SessionStatsSubscription": {
            "SessionID": sessionID,
        },
    }
}