import {v4 as uuid} from 'uuid';
import React, {useEffect, useState} from 'react';
import useWebSocket, {ReadyState} from 'react-use-websocket';
import {useCookies} from 'react-cookie';
import {Link, Outlet, useParams, useRoutes} from "react-router-dom";
import QRCode from "react-qr-code";
import {CreateGameCMD, JoinGameCMD, MoveCMD, StartGameCMD} from "./commands";


export function App() {
    const routes = [
        {
            element: <Layout/>,
            children: [
                {
                    index: true,
                    element: <Info/>,
                }, {
                    path: "game/:gameID",
                    element: <Game/>,
                }
            ]
        }
    ]

    let element = useRoutes(routes);

    return (
        <div>
            <h1>Route Objects Example</h1>
            {element}
        </div>
    )
}

export function Layout() {
    return (
        <div className="main">
            <h1>Gra</h1>
            <main>
                <Outlet/>
            </main>
            <aside>
                <Link className="button-29"
                      to={"game/" + uuid()}>
                    StartGame
                </Link>
            </aside>
        </div>
    );
}


export function Info() {
    return (
        <div>
            <h2>Info</h2>
            <Link className="button-29"
                  to={"game/" + uuid()}>
                StartGame
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
                    style = "X"
                } else if (player === SecondPlayerID) {
                    style = "O"
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


function serverURL(gameID) {
    return 'ws://' + document.location.hostname + ':8080/play/' + gameID
}

function gameURL(gameID) {
    return 'http://' + document.location.hostname + ':3000/#/game/' + gameID
}

export function Game() {
    let {gameID} = useParams();

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
        setSocketUrl(serverURL(gameID));
    }, [gameID]);

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
            // Automatically create a game if the playerID is already set
            sendJsonMessage(CreateGameCMD(cookies.playerID))
        } else if (currentGameState?.GameWaitingForPlayer) {
            if (currentGameState?.GameWaitingForPlayer?.FirstPlayerID !== cookies.playerID) {
                // Automatically join the game if the playerID is already set
                sendJsonMessage(JoinGameCMD(cookies.playerID))
            }
        }
    }, [readyState, currentGameState, cookies, sendJsonMessage]);

    return (
        <div className="game">
            <div className="game-board">
                <Board state={currentGameState} transition={sendJsonMessage} playerID={cookies.playerID}/>
            </div>
            <div className="game-info">
                <p>Player: {cookies.playerID}</p>
                <p>GameID: {gameID}</p>
                <Actions state={currentGameState} transition={sendJsonMessage} playerID={cookies.playerID}/>
                <p>Game server is currently {connectionStatus}</p>
                <QRCode value={gameURL(gameID)}
                        style={{height: "auto", maxWidth: "200px", width: "100%"}}/>
            </div>
        </div>
    );
}

function Actions({state, transition, playerID}) {
    if (state?.GameWaitingForPlayer) {
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
                    Make a move! ⚽️
                </div>
            )
        } else {
            return (
                <div>
                    Waiting for the other player to make a move... ⏳
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
                            onClick={() => window.location.reload()}>Play again
                    </button>
                </div>
            )
        } else {
            return (
                <div>
                    You lost! 😢
                    <button className="button-29"
                            onClick={() => window.location.reload()}>Play again
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
                        onClick={() => window.location.reload()}>Play again
                </button>
            </div>
        )
    } else {
        return (
            <div>
                <button className="button-29"
                        onClick={() => transition(StartGameCMD("x", "o"))}>
                    Start Game
                </button>
                <button className="button-29"
                        onClick={() => transition(CreateGameCMD("x"))}>
                    Create Game
                </button>
            </div>
        )
    }
}