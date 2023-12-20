package util

const (
	ConnHost                   = "192.168.0.148"
	ConnPort                   = "8080"
	ConnType                   = "tcp"
	MaxDataLen                 = 128
	MaxMsgDataLen              = 4
	MaxClients                 = 4
	ArgSep                     = ";" //argument separator in messages
	MaxInvalidOp               = 5   //max number of invalid operations before disconnecting client
	PingTime                   = 3   //time between pings
	MaxNoPingReceived          = 3   //if 3 pings are not received, client is disconnected
	MaxSecondsBeforeDisconnect = 80  //time before completely disconnecting client, must be bigger than PingTime*MaxNoPingReceived

	//Size of message (msgdatalen) is a 4 digit number -> 0 ... 9999 bytes
	// maxmsdgdatalen says how large data part is
	MsgHeaderLen = len(MsgMagic) + len(MsgLoginOpcode) + MaxMsgDataLen

	//magic word
	MsgMagic = "KIVUPS" //magic word needed

	//Login operation arguments: string, client response is OK and board size or ERR
	MsgLoginOpcode = "001"

	//Join operation has no arguments, client response is OK or ERR
	MsgJoinOpcode = "002"

	//Move operation arguments: int;int, client response contains board in parsable format
	MsgMoveOpcode = "003"

	//Operation play again has no arguments
	MsgPlayAgainOpcode = "004"

	//Game started has no arguments, client response contains name of the other player
	MsgGameStartedOpcode = "005"

	//Return to start has no arguments, returns OK but returns ERR and GameGone if game does not exist anymore
	MsgReturnToStartOpcode = "006"

	//Server doesnt receive this, only sends it to client with winner name (or draw)
	MsgGameOverOpcode = "007"

	//unused
	MsgOkOpcode = "008"

	//unused
	MsgErrOpcode = "009"

	//Server doesnt receive this, only sends it to client signifying that it is his turn
	MsgYourTurnOpcode = "010"

	//Ping operation has no arguments, client response is OK (should be)
	MsgPingOpcode = "011"

	//Recovery operation has no arguments, client response is recovery state depending on state of player and game (board, winner, etc.)
	MsgRecoveryOpcode = "012"

	//Server doesnt receive this, only sends it to client signifying that it should pause the game (because other player disconnected)
	MsgPauseOpcode = "013"

	//Server doesnt receive this, only sends it to client signifying that it should continue the game (because other player reconnected)
	MsgContinueOpcode = "014"

	//send some status info to client
	MsgStatusOpcode = "015"
)

// info for client that their msg was not valid and the server didnt like it so it will kick them if they keep sending invalid msgs
const (
	SrvErrInvalidOp = "criticalerror"
)

// extra info (data) for opcodes (client messages)
const (
	ClientMsgGameGone = "gamegone"
	ClientMsgOk       = "ok"
	ClientMsgErr      = "err"

	//recovery
	ClientMsgRecoveryLogin         = "recovery_login"
	ClientMsgRecovery_InLobby      = "recovery_inlobby"
	ClientMsgRecovery_ReadyForGame = "recovery_readyforgame"

	//ClientMsgRecovery_InGame                = "recovery_ingame"
	ClientMsgRecovery_InGame_YourTurn       = "recovery_ingame_yourturn"
	ClientMsgRecovery_InGame_OtherTurn      = "recovery_ingame_otherturn"
	ClientMsgRecovery_InGame_GameGone       = "recovery_ingame_gamegone"
	ClientMsgRecovery_InGame_OtherPlayAgain = "recovery_ingame_otherplayagain"
	ClientMsgRecovery_InGame_GameOver       = "recovery_ingame_gameover"
)

// client staus
const (
	InLobby      = 1
	InGame       = 2
	ReadyForGame = 3
)

// Game constants
const (
	colSep           = "|"
	rowSep           = "--"
	defaultBoardSize = 3
	//Game state
	WaitingForPlayersReady  = 1
	WaitingForPlayerOneMove = 2
	WaitingForPlayerTwoMove = 3
	GameOver                = 4
	//Game over state
	NotOver      = 5
	PlayerOneWin = 6
	PlayerTwoWin = 7
	Draw         = 8
)
