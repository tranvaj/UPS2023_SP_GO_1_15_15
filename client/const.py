# Connection constants
ConnHost = "192.168.0.148"
ConnPort = 8080
ConnType = "tcp"
MaxDataLen = 128
MaxMsgDataLen = 4
MaxClients = 2
ArgSep = ";"
PingTime          = 3
MaxNoPingReceived = 3
RecoveryMaxAttempts = 5

# Message constants
MsgMagic = "KIVUPS"
MsgLoginOpcode = "001"
MsgJoinOpcode = "002"
MsgMoveOpcode = "003"
MsgPlayAgainOpcode = "004"

#Broadcast messages for clients
MsgGameStartedOpcode = "005"
MsgReturnToStartOpcode = "006"
MsgGameOverOpcode = "007"
MsgYourTurnOpcode = "010"


MsgOkOpcode = "008"
MsgErrOpcode = "009"

MsgPingOpcode = "011"
MsgRecoveryOpcode = "012"
MsgPauseOpcode = "013"
MsgContinueOpcode = "014"
MsgStatusOpcode = "015"

#MsgOtherPlayerDisconnectedOpcode = "014"


ClientMsgGameGone = "gamegone"
ClientMsgOk       = "ok"
ClientMsgErr      = "err"

ClientMsgRecoveryLogin         = "recovery_login"
ClientMsgRecovery_InLobby      = "recovery_inlobby"
ClientMsgRecovery_ReadyForGame = "recovery_readyforgame"

ClientMsgRecovery_InGame_YourTurn       = "recovery_ingame_yourturn"
ClientMsgRecovery_InGame_OtherTurn      = "recovery_ingame_otherturn"
ClientMsgRecovery_InGame_GameGone       = "recovery_ingame_gamegone"
ClientMsgRecovery_InGame_OtherPlayAgain = "recovery_ingame_otherplayagain"
ClientMsgRecovery_InGame_GameOver       = "recovery_ingame_gameover"



MsgHeaderLen = len(MsgMagic) + len(MsgLoginOpcode) + MaxMsgDataLen


# Game constants
colSep           = "|"
rowSep           = "--"
defaultBoardSize = 5
WaitingForPlayersReady = 1
WaitingForPlayerOneMove = 2
WaitingForPlayerTwoMove = 3
GameOver = 4
NotOver = 5
PlayerOneWin = 6
PlayerTwoWin = 7
Draw = 8
