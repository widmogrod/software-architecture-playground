import {v4 as uuid} from 'uuid';
import React, {useEffect, useState} from 'react';
import useWebSocket, {ReadyState} from 'react-use-websocket';
import {useCookies} from 'react-cookie';
import {Link, Outlet, useParams, useRoutes} from "react-router-dom";
import QRCode from "react-qr-code";
import {JoinGameCMD, MoveCMD, StartGameCMD} from "./cmd.game";
import * as manage from "./cmd.manage";

/*
* TODO
*  - [√] When game is over, Manager transition to SessionReady, but this remove from UI the winning sequence!
*  - [√] Add play again that use NewGameCMD()
*  - [ ] Inform when players gets disconnected, to not show that "waiting for other player" forever
*  - [ ] Consider auto-start of the game in Manager, not by sending StartGameCMD from UI
*  - [ ] Consider timout for waiting for player moves
*  - [ ] List all sessions that are waiting for players, alow to join them, show when session expires was active
*  - [ ] Add bigger board
*/

export function App() {
    const routes = [
        {
            element: <Layout/>,
            children: [
                {
                    index: true,
                    element: <Info/>,
                }, {
                    path: "/game/:sessionID",
                    element: <Game/>,
                }
            ]
        }
    ]

    return useRoutes(routes);
}

export function Layout() {
    return (
        <div>
            <Outlet/>
        </div>
    );
}


export function Info() {
    return (
        <div className="nav">
            <Link className="button-29"
                  to={"game/" + uuid()}>
                Create game session
            </Link>
        </div>
    );
}

function Square({value, isWin, onSquareClick}) {
    let className = isWin ? "square win" : "square"
    return (
        <button className={className} onClick={onSquareClick}>
            {value}
        </button>
    );
}


function Board({state, transition, playerID}) {
    let {MovesTaken, WiningSequence, FirstPlayerID, SecondPlayerID} = (() => {
        if (state?.GameProgress) {
            return state?.GameProgress
        } else if (state?.GameEndWithWin) {
            return state?.GameEndWithWin
        } else if (state?.GameEndWithDraw) {
            return state?.GameEndWithDraw
        }

        return {}
    })()

    let result = []
    for (let i = 1; i <= 3; i++) {
        let row = []
        for (let j = 1; j <= 3; j++) {
            let move = "" + i + "." + j

            let isWin = WiningSequence?.find((m) => m === move)

            let style = ""
            let player = MovesTaken?.[move]
            if (player) {
                if (player === FirstPlayerID) {
                    style = "🧜‍"
                } else if (player === SecondPlayerID) {
                    style = "🖤"
                }
            }

            row.push(<Square key={move} value={style} isWin={isWin}
                             onSquareClick={() => transition(MoveCMD(playerID, move))}/>)
        }

        result.push(<div key={i} className="board-row">{row}</div>)
    }

    return (
        <>
            {result}
        </>
    );
}


function serverURL(sessionID) {
    return 'ws://' + document.location.hostname + ':8080/play/' + sessionID
}

function gameURL(sessionID) {
    return 'http://' + document.location.hostname + ':3000/#/game/' + sessionID
}

export function Game() {
    let {sessionID} = useParams();

    const [squareStyle, setSquareStyle] = useState(["🧜‍", "🖤"]);
    const [playerNo, setPlayerNo] = useState(0)

    const [currentGameState, setGameState] = useState({});
    const [cookies, setCookie] = useCookies(['playerID']);

    const [socketUrl, setSocketUrl] = useState(null)

    const {sendJsonMessage, lastJsonMessage, readyState} = useWebSocket(socketUrl);

    const connectionStatus = {
        [ReadyState.CONNECTING]: '🌀Connecting',
        [ReadyState.OPEN]: '✅Open',
        [ReadyState.CLOSING]: '😱Closing',
        [ReadyState.CLOSED]: '❌Closed',
        [ReadyState.UNINSTANTIATED]: '❓Uninstantiated',
    }[readyState];

    if (!cookies.playerID) {
        setCookie('playerID', uuid(), {path: '/'});
    }

    useEffect(() => {
        setSocketUrl(serverURL(sessionID));
    }, [sessionID]);

    useEffect(() => {
        setGameState(lastJsonMessage)
    }, [lastJsonMessage]);

    useEffect(() => {
        if (readyState !== ReadyState.OPEN) {
            return
        }
        if (!cookies.playerID) {
            return
        }
        if (!currentGameState) {
            sendJsonMessage(manage.CreateSessionCMD(sessionID, 2))
            sendJsonMessage(manage.JoinGameSessionCMD(sessionID, cookies.playerID))
            // Automatically create a game if the playerID is already set
            // sendJsonMessage(CreateGameCMD(cookies.playerID))
        } else if (currentGameState?.SessionWaitingForPlayers) {
            if (!currentGameState?.SessionWaitingForPlayers?.Players.find((x) => x == cookies.playerID)) {
                // Automatically join the game if the playerID is already set
                // sendJsonMessage(JoinGameCMD(cookies.playerID))
                sendJsonMessage(manage.JoinGameSessionCMD(sessionID, cookies.playerID))
            }
        } else if (currentGameState?.SessionReady) {
            currentGameState?.SessionReady?.Players.find((x, i) => {
                if (x == cookies.playerID) {
                    setPlayerNo(i)
                }
            })
            if (currentGameState?.SessionReady?.Players[0] == cookies.playerID) {
                let gameId = uuid()
                sendJsonMessage(manage.NewGameCMD(sessionID, gameId))
                sendJsonMessage(manage.GameActionCMD(sessionID, gameId, StartGameCMD(
                    currentGameState?.SessionReady?.Players[0],
                    currentGameState?.SessionReady?.Players[1]
                )))
            }
        }

    }, [readyState, currentGameState, cookies, sendJsonMessage]);


    function transition(cmd) {
        let gid = currentGameState?.SessionInGame?.GameID
        sendJsonMessage(manage.GameActionCMD(sessionID, gid, cmd))
    }

    function newGame() {
        let gameId = uuid()
        sendJsonMessage(manage.NewGameCMD(sessionID, gameId))
        sendJsonMessage(manage.GameActionCMD(sessionID, gameId, StartGameCMD(
            currentGameState?.SessionInGame?.Players[0],
            currentGameState?.SessionInGame?.Players[1],
        )))
    }

    return (
        <>
            <div className="nav">
                <Link className="button-29"
                      to={"/"}>
                    👈
                </Link>
            </div>
            <div className="game">
                <div className="game-info">
                    <p>You are {squareStyle[playerNo]} <b>vs</b> other {squareStyle.filter((_, i) => i != playerNo)}</p>
                    <Actions state={currentGameState?.SessionInGame?.GameState}
                             transition={transition}
                             newGame={newGame}
                             playerID={cookies.playerID}/>
                </div>
                <div className="game-board">
                    <Board state={currentGameState?.SessionInGame?.GameState} transition={transition}
                           playerID={cookies.playerID}/>
                </div>
                <div className="game-share">
                    <p>Ask friend to scan code to join</p>
                    <QRCode value={gameURL(sessionID)}
                            style={{height: "auto", maxWidth: "200px", width: "100%"}}/>

                </div>
                <div className="game-debug">
                    <p>Player: {cookies.playerID}</p>
                    <p>SessionID: {sessionID}</p>
                    <p>GameID: {currentGameState?.SessionInGame?.GameID}</p>
                    <p>Game server is currently {connectionStatus}</p>
                    <p><code>{currentGameState?.SessionInGame?.GameProblem}</code></p>
                </div>
            </div>
        </>
    );
}

function Actions({state, transition, playerID, newGame}) {
    if (state?.GameWaitingForPlayers) {
        return (
            <div>
                <button className="button-29"
                        onClick={() => transition(JoinGameCMD())}>
                    Invite Player
                </button>
            </div>
        )
    } else if (state?.GameProgress) {
        let {NextMovePlayerID} = state?.GameProgress
        if (NextMovePlayerID === playerID) {
            return (
                <div>
                    <b>Your a move! ⚽️</b>
                </div>
            )
        } else {
            return (
                <div>
                    <b>⏳ Other player is moving </b>
                </div>
            )
        }

    } else if (state?.GameEndWithWin) {
        let {Winner} = state?.GameEndWithWin
        if (Winner === playerID) {
            return (
                <div>
                    You won! 🎉
                    <button className="button-29"
                            onClick={() => newGame()}>Play again
                    </button>
                </div>
            )
        } else {
            return (
                <div>
                    You lost! 😢
                    <button className="button-29"
                            onClick={() => newGame()}>Play again
                    </button>
                </div>
            )
        }
    } else if (state?.GameEndWithDraw) {
        return (
            <div>
                Game result is a <b>DRAW!</b> 🤝
                <br/>
                <button className="button-29"
                        onClick={() => newGame()}>Play again
                </button>
            </div>
        )
    }
}