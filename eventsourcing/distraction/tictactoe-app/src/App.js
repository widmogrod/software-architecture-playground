import {v4 as uuid} from 'uuid';
import React, {useEffect, useState} from 'react';
import useWebSocket, {ReadyState} from 'react-use-websocket';
import {useCookies} from 'react-cookie';
import {Link, Outlet, useNavigate, useParams, useRoutes} from "react-router-dom";
import QRCode from "react-qr-code";
import {GiveUpCMD, MoveCMD, StartGameCMD} from "./cmd.game";
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
*  - [ ] Add remote play, local play, and play with AI
*  - [ ] Player who clicks "Play again" should be the first to move
*  - [√] Remove AvailableMoves! they are not needed, and make code harder to read
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
                    path: "/game/:sessionID/:widthAndHeight?/:lengthToWin?",
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
        <div className="game-options">
            <div className="game-option">
                <h2><b>Classic</b> tic-tac-toe</h2>
                <p>Classic 3 in a row wins. Played on board with dimensions 3x3</p>
                <Board2 cols={3} rows={3}
                        winingSequence={["1.1", "2.2", "3.3"]}/>
                <Link className={"button-action"} to={"game/" + uuid()}>
                    Play
                </Link>
            </div>
            <div className="game-option">
                <h2>4 in a row</h2>
                <p>Much more demanding variant of tic tac toe. Board has 8x8 dimensions, and four in row wins </p>
                <Board2 cols={8} rows={8}
                        winingSequence={["3.3", "4.4", "5.5", "6.6"]}/>
                <Link className={"button-action"} to={"game/" + uuid() + "/8/4"}>
                    Play
                </Link>
            </div>
        </div>
    );
}

function Square({value, isWin, onSquareClick}) {
    let className = isWin ? "square win" : "square"
    return (
        <button className={className}
                onClick={onSquareClick}>
            {value}
        </button>
    );
}


function Board({state, transition, playerID, squareStyle}) {
    let {
        MovesTaken,
        WiningSequence,
        TicTacToeBaseState,
    } = (() => {
        if (state?.GameProgress) {
            return state?.GameProgress
        } else if (state?.GameEndWithWin) {
            return state?.GameEndWithWin
        } else if (state?.GameEndWithDraw) {
            return state?.GameEndWithDraw
        }

        return {}
    })()

    let {
        FirstPlayerID, SecondPlayerID,
        BoardRows, BoardCols
    } = TicTacToeBaseState || {}

    return (
        <Board2 rows={BoardRows}
                cols={BoardCols}
                winingSequence={WiningSequence}
                movesTaken={MovesTaken}
                playersStyle={({playerID}) => {
                    switch (playerID) {
                        case FirstPlayerID:
                            return squareStyle[0]
                        case SecondPlayerID:
                            return squareStyle[1]
                        default:
                            return ""
                    }
                }}
                onSquareClick={({move}) => transition(MoveCMD(playerID, move))}/>
    );
}

function Board2({movesTaken, playersStyle, rows, cols, winingSequence, onSquareClick}) {
    let result = []
    for (let i = 1; i <= rows; i++) {
        let row = []
        for (let j = 1; j <= cols; j++) {
            let move = "" + i + "." + j
            let isWin = winingSequence?.find((m) => m === move)

            let playerID = movesTaken?.[move]
            let style = playersStyle?.({playerID})

            row.push(
                <td key={move}>
                    <div className={"cell"}>
                        <Square value={style} isWin={isWin}
                                onSquareClick={() => onSquareClick?.({move})}
                        />
                    </div>
                </td>
            )
        }

        result.push(<tr key={i}>{row}</tr>)
    }

    return (
        <table className={`tictactoe size${rows}x${cols}`}>
            <tbody>
            {result}
            </tbody>
        </table>
    )
}


function serverURL(sessionID) {
    return 'wss://al0ofi3lke.execute-api.eu-west-1.amazonaws.com/dev'
    // return 'ws://' + document.location.hostname + ':8080/play/' + sessionID
}

function gameURL(sessionID) {
    return 'https://' + document.location.hostname + '/#/game/' + sessionID
}

export function Game() {
    const navigate = useNavigate();
    let {sessionID, widthAndHeight, lengthToWin} = useParams();

    const [squareStyle] = useState(["🧜‍", "🖤"]);
    const [settings, setSettings] = useState({
        WidthAndHeight: widthAndHeight,
        LengthToWin: lengthToWin
    });
    const [playerNo, setPlayerNo] = useState(0)

    const [currentGameState, setGameState] = useState({});
    const [cookies, setCookie] = useCookies(['playerID']);

    const [socketUrl, setSocketUrl] = useState(null)

    const {sendJsonMessage, lastJsonMessage, readyState} = useWebSocket(socketUrl, {
        reconnectAttempts: 10,
        reconnectInterval: 3000,
    });

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
            sendJsonMessage(manage.SequenceCMD([
                manage.CreateSessionCMD(sessionID, 2),
                manage.JoinGameSessionCMD(sessionID, cookies.playerID)
            ]))
            // Automatically create a game if the playerID is already set
            // sendJsonMessage(CreateGameCMD(cookies.playerID))
        } else if (currentGameState?.SessionWaitingForPlayers) {
            if (!currentGameState?.SessionWaitingForPlayers?.Players.find((x) => x === cookies.playerID)) {
                // Automatically join the game if the playerID is already set
                // sendJsonMessage(JoinGameCMD(cookies.playerID))
                sendJsonMessage(manage.JoinGameSessionCMD(sessionID, cookies.playerID))
            }
        } else if (currentGameState?.SessionReady) {
            currentGameState?.SessionReady?.Players.forEach((x, i) => {
                if (x === cookies.playerID) {
                    setPlayerNo(i)
                }
            })
            if (currentGameState?.SessionReady?.Players[0] === cookies.playerID) {
                newGame(widthAndHeight, lengthToWin)
            }
        }

        //eslint-disable-next-line
    }, [readyState, currentGameState, cookies, sendJsonMessage]);

    const setProperty = (obj, path, value) => {
        const [head, ...rest] = path.split('#')

        return {
            ...obj,
            [head]: rest.length
                ? setProperty(obj[head], rest.join('#'), value)
                : value
        }
    }

    function transition(cmd) {

        // this is a hack for optimistic concurrency
        if (currentGameState.SessionInGame?.GameState?.GameProgress?.MovesTaken && cmd.MoveCMD?.Position) {
            setGameState(setProperty(
                currentGameState,
                "SessionInGame#GameState#GameProgress#MovesTaken#" + cmd.MoveCMD?.Position,
                cmd.MoveCMD?.PlayerID
            ))
        }

        let gid = currentGameState?.SessionInGame?.GameID
        sendJsonMessage(manage.GameActionCMD(sessionID, gid, cmd))
    }

    function playAgain() {
        newGame(settings.WidthAndHeight, settings.LengthToWin)
    }

    function newGame(wh, l) {
        let gameId = uuid()

        setSettings({
            WidthAndHeight: wh,
            LengthToWin: l,
        })

        sendJsonMessage(manage.SequenceCMD([
            manage.NewGameCMD(sessionID, gameId),
            manage.GameActionCMD(sessionID, gameId, StartGameCMD(
                "",
                "",
                wh,
                l,
            )),
        ]))
    }

    function giveUpOrBack() {
        if (currentGameState?.SessionInGame?.GameState?.GameProgress) {
            transition(GiveUpCMD(cookies.playerID))
        } else {
            navigate("/")
        }
    }

    function playWithBot() {
        sendJsonMessage(manage.GameSessionWithBotCMD(sessionID))
        // playAgain()
    }


    function playLocally() {
       alert("Not implemented yet")
    }

    return (
        <>
            <ul className="nav">
                <li>
                    <button className="button-close" onClick={giveUpOrBack}></button>
                </li>
                <li>You are {squareStyle[playerNo]} <b>vs</b> {squareStyle.filter((_, i) => i !== playerNo)}</li>
            </ul>
            <div className="game">
                <div className="game-info">
                    <Actions state={currentGameState?.SessionInGame?.GameState}
                             transition={transition}
                             newGame={newGame}
                             playAgain={playAgain}
                             playerID={cookies.playerID}/>
                </div>
                <div className="game-board">
                    <Board state={currentGameState?.SessionInGame?.GameState}
                           transition={transition}
                           squareStyle={squareStyle}
                           playerID={cookies.playerID}/>
                </div>
                {currentGameState?.SessionWaitingForPlayers &&
                    <div className="game-share">
                        <p>
                            Ask friend to scan code to join the game
                        </p>
                        <p>
                            <QRCode value={gameURL(sessionID)}
                                    style={{height: "auto", maxWidth: "200px", width: "100%"}}/>
                        </p>
                        <p>
                            or play <button className="button-text" onClick={() => playWithBot()}>with bot 🤖 </button>
                            &nbsp;or <button className="button-text" onClick={() => playLocally()}> locally</button> on device
                        </p>

                    </div>}
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

function ChangeGameActions({newGame}) {
    return (
        <>
            <button className="button-text"
                    onClick={() => newGame(3, 3)}>
                3x3
            </button>
            <span> or </span>
            <button className="button-text"
                    onClick={() => newGame(8, 4)}>
                4x8
            </button>
        </>
    )
}

function PostGameActions({newGame, playAgain}) {
    return (
        <p>
            <button className="button-action"
                    onClick={() => playAgain()}>Play again
            </button>
            <br/>
            <span> or change game </span>
            <br/>
            <ChangeGameActions newGame={newGame}/>
        </p>
    )
}

function Actions({state, playerID, newGame, playAgain}) {
    if (state?.GameProgress) {
        let callToAction
        let {NextMovePlayerID} = state?.GameProgress
        if (NextMovePlayerID === playerID) {
            callToAction = <span>⚽️ Your <b> move</b>...</span>
        } else {
            callToAction = <span>⏳ <b>Wait</b> on the other player move... </span>
        }

        return (
            <p>
                {callToAction}
            </p>
        )

    } else if (state?.GameEndWithWin) {
        let {Winner} = state?.GameEndWithWin
        if (Winner === playerID) {
            return (
                <div>
                    <p>Bravo! <b>you won!</b> 🎉🎉🎉</p>
                    <PostGameActions newGame={newGame} playAgain={playAgain}/>
                </div>
            )
        } else {
            return (
                <div>
                    <p>You lost! 😢. Try <b>again</b></p>
                    <PostGameActions newGame={newGame} playAgain={playAgain}/>
                </div>
            )
        }
    } else if (state?.GameEndWithDraw) {
        return (
            <div>
                <p><b>DRAW!</b> 🤝. Good game!</p>
                <PostGameActions newGame={newGame} playAgain={playAgain}/>
            </div>
        )

    } else if (state === null) {
        return (
            <div>
                <p>Taking to long?</p>
                <button className="button-action"
                        onClick={() => playAgain()}>Play
                </button>
            </div>
        )
    }
}