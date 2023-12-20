package util

const (
	ConnHost                   = "192.168.0.148"
	ConnPort                   = "8080"
	ConnType                   = "tcp"
	MaxDataLen                 = 128
	MaxMsgDataLen              = 4
	MaxClients                 = 4
	ArgSep                     = ";"
	MaxInvalidOp               = 5
	PingTime                   = 3
	MaxNoPingReceived          = 3  //if 3 pings are not received, client is disconnected
	MaxSecondsBeforeDisconnect = 80 //time before completely disconnecting client, must be bigger than PingTime*MaxNoPingReceived

	//Size of message (msgdatalen) is a 4 digit number -> 0 ... 9999 bytes
	// maxmsdgdatalen says how large data part is
	MsgHeaderLen = len(MsgMagic) + len(MsgLoginOpcode) + MaxMsgDataLen

	//magic word
	MsgMagic = "KIVUPS" //magic word needed

	//Login operation arguments: string
	MsgLoginOpcode = "001"

	//Join operation has no arguments
	MsgJoinOpcode = "002"

	//Move operation arguments: int;int
	MsgMoveOpcode = "003"

	//Operation play again has no arguments
	MsgPlayAgainOpcode = "004"

	MsgGameStartedOpcode = "005"

	MsgReturnToStartOpcode = "006"

	MsgGameOverOpcode = "007"

	MsgOkOpcode = "008"

	MsgErrOpcode = "009"

	MsgYourTurnOpcode = "010"

	MsgPingOpcode = "011"

	MsgRecoveryOpcode = "012"

	MsgPauseOpcode = "013"

	MsgContinueOpcode = "014"

	MsgStatusOpcode = "015"
)

// info for server
const (
	SrvErrInvalidOp = "criticalerror"
)

// client messages
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
