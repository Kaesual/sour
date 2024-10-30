package protocol

import (
	"fmt"
)

const PROTOCOL_VERSION = 260

type MessageCode int32 // network message code

const N_EMPTY MessageCode = -1

const (
	N_CONNECT MessageCode = iota
	N_SERVINFO
	N_WELCOME
	N_INITCLIENT
	N_POS
	N_TEXT
	N_SOUND
	N_CDIS
	N_SHOOT
	N_EXPLODE
	N_SUICIDE
	N_DIED
	N_DAMAGE
	N_HITPUSH
	N_SHOTFX
	N_EXPLODEFX
	N_TRYSPAWN
	N_SPAWNSTATE
	N_SPAWN
	N_FORCEDEATH
	N_GUNSELECT
	N_TAUNT
	N_MAPCHANGE
	N_MAPVOTE
	N_TEAMINFO
	N_ITEMSPAWN
	N_ITEMPICKUP
	N_ITEMACC
	N_TELEPORT
	N_JUMPPAD
	N_PING
	N_PONG
	N_CLIENTPING
	N_TIMEUP
	N_FORCEINTERMISSION
	N_SERVMSG
	N_ITEMLIST
	N_RESUME
	N_EDITMODE
	N_EDITENT
	N_EDITF
	N_EDITT
	N_EDITM
	N_FLIP
	N_COPY
	N_PASTE
	N_ROTATE
	N_REPLACE
	N_DELCUBE
	N_REMIP
	N_EDITVSLOT
	N_UNDO
	N_REDO
	N_NEWMAP
	N_GETMAP
	N_SENDMAP
	N_CLIPBOARD
	N_EDITVAR
	N_MASTERMODE
	N_KICK
	N_CLEARBANS
	N_CURRENTMASTER
	N_SPECTATOR
	N_SETMASTER
	N_SETTEAM
	N_BASES
	N_BASEINFO
	N_BASESCORE
	N_REPAMMO
	N_BASEREGEN
	N_ANNOUNCE
	N_LISTDEMOS
	N_SENDDEMOLIST
	N_GETDEMO
	N_SENDDEMO
	N_DEMOPLAYBACK
	N_RECORDDEMO
	N_STOPDEMO
	N_CLEARDEMOS
	N_TAKEFLAG
	N_RETURNFLAG
	N_RESETFLAG
	N_INVISFLAG
	N_TRYDROPFLAG
	N_DROPFLAG
	N_SCOREFLAG
	N_INITFLAGS
	N_SAYTEAM
	N_CLIENT
	N_AUTHTRY
	N_AUTHKICK
	N_AUTHCHAL
	N_AUTHANS
	N_REQAUTH
	N_PAUSEGAME
	N_GAMESPEED
	N_ADDBOT
	N_DELBOT
	N_INITAI
	N_FROMAI
	N_BOTLIMIT
	N_BOTBALANCE
	N_MAPCRC
	N_CHECKMAPS
	N_SWITCHNAME
	N_SWITCHMODEL
	N_SWITCHTEAM
	N_INITTOKENS
	N_TAKETOKEN
	N_EXPIRETOKENS
	N_DROPTOKENS
	N_DEPOSITTOKENS
	N_STEALTOKENS
	N_SERVCMD
	N_DEMOPACKET
	NUMMSG
)

func (e MessageCode) String() string {
	switch e {
	case N_CONNECT:
		return "N_CONNECT"
	case N_SERVINFO:
		return "N_SERVINFO"
	case N_WELCOME:
		return "N_WELCOME"
	case N_INITCLIENT:
		return "N_INITCLIENT"
	case N_POS:
		return "N_POS"
	case N_TEXT:
		return "N_TEXT"
	case N_SOUND:
		return "N_SOUND"
	case N_CDIS:
		return "N_CDIS"
	case N_SHOOT:
		return "N_SHOOT"
	case N_EXPLODE:
		return "N_EXPLODE"
	case N_SUICIDE:
		return "N_SUICIDE"
	case N_DIED:
		return "N_DIED"
	case N_DAMAGE:
		return "N_DAMAGE"
	case N_HITPUSH:
		return "N_HITPUSH"
	case N_SHOTFX:
		return "N_SHOTFX"
	case N_EXPLODEFX:
		return "N_EXPLODEFX"
	case N_TRYSPAWN:
		return "N_TRYSPAWN"
	case N_SPAWNSTATE:
		return "N_SPAWNSTATE"
	case N_SPAWN:
		return "N_SPAWN"
	case N_FORCEDEATH:
		return "N_FORCEDEATH"
	case N_GUNSELECT:
		return "N_GUNSELECT"
	case N_TAUNT:
		return "N_TAUNT"
	case N_MAPCHANGE:
		return "N_MAPCHANGE"
	case N_MAPVOTE:
		return "N_MAPVOTE"
	case N_TEAMINFO:
		return "N_TEAMINFO"
	case N_ITEMSPAWN:
		return "N_ITEMSPAWN"
	case N_ITEMPICKUP:
		return "N_ITEMPICKUP"
	case N_ITEMACC:
		return "N_ITEMACC"
	case N_TELEPORT:
		return "N_TELEPORT"
	case N_JUMPPAD:
		return "N_JUMPPAD"
	case N_PING:
		return "N_PING"
	case N_PONG:
		return "N_PONG"
	case N_CLIENTPING:
		return "N_CLIENTPING"
	case N_TIMEUP:
		return "N_TIMEUP"
	case N_FORCEINTERMISSION:
		return "N_FORCEINTERMISSION"
	case N_SERVMSG:
		return "N_SERVMSG"
	case N_ITEMLIST:
		return "N_ITEMLIST"
	case N_RESUME:
		return "N_RESUME"
	case N_EDITMODE:
		return "N_EDITMODE"
	case N_EDITENT:
		return "N_EDITENT"
	case N_EDITF:
		return "N_EDITF"
	case N_EDITT:
		return "N_EDITT"
	case N_EDITM:
		return "N_EDITM"
	case N_FLIP:
		return "N_FLIP"
	case N_COPY:
		return "N_COPY"
	case N_PASTE:
		return "N_PASTE"
	case N_ROTATE:
		return "N_ROTATE"
	case N_REPLACE:
		return "N_REPLACE"
	case N_DELCUBE:
		return "N_DELCUBE"
	case N_REMIP:
		return "N_REMIP"
	case N_EDITVSLOT:
		return "N_EDITVSLOT"
	case N_UNDO:
		return "N_UNDO"
	case N_REDO:
		return "N_REDO"
	case N_NEWMAP:
		return "N_NEWMAP"
	case N_GETMAP:
		return "N_GETMAP"
	case N_SENDMAP:
		return "N_SENDMAP"
	case N_CLIPBOARD:
		return "N_CLIPBOARD"
	case N_EDITVAR:
		return "N_EDITVAR"
	case N_MASTERMODE:
		return "N_MASTERMODE"
	case N_KICK:
		return "N_KICK"
	case N_CLEARBANS:
		return "N_CLEARBANS"
	case N_CURRENTMASTER:
		return "N_CURRENTMASTER"
	case N_SPECTATOR:
		return "N_SPECTATOR"
	case N_SETMASTER:
		return "N_SETMASTER"
	case N_SETTEAM:
		return "N_SETTEAM"
	case N_BASES:
		return "N_BASES"
	case N_BASEINFO:
		return "N_BASEINFO"
	case N_BASESCORE:
		return "N_BASESCORE"
	case N_REPAMMO:
		return "N_REPAMMO"
	case N_BASEREGEN:
		return "N_BASEREGEN"
	case N_ANNOUNCE:
		return "N_ANNOUNCE"
	case N_LISTDEMOS:
		return "N_LISTDEMOS"
	case N_SENDDEMOLIST:
		return "N_SENDDEMOLIST"
	case N_GETDEMO:
		return "N_GETDEMO"
	case N_SENDDEMO:
		return "N_SENDDEMO"
	case N_DEMOPLAYBACK:
		return "N_DEMOPLAYBACK"
	case N_RECORDDEMO:
		return "N_RECORDDEMO"
	case N_STOPDEMO:
		return "N_STOPDEMO"
	case N_CLEARDEMOS:
		return "N_CLEARDEMOS"
	case N_TAKEFLAG:
		return "N_TAKEFLAG"
	case N_RETURNFLAG:
		return "N_RETURNFLAG"
	case N_RESETFLAG:
		return "N_RESETFLAG"
	case N_INVISFLAG:
		return "N_INVISFLAG"
	case N_TRYDROPFLAG:
		return "N_TRYDROPFLAG"
	case N_DROPFLAG:
		return "N_DROPFLAG"
	case N_SCOREFLAG:
		return "N_SCOREFLAG"
	case N_INITFLAGS:
		return "N_INITFLAGS"
	case N_SAYTEAM:
		return "N_SAYTEAM"
	case N_CLIENT:
		return "N_CLIENT"
	case N_AUTHTRY:
		return "N_AUTHTRY"
	case N_AUTHKICK:
		return "N_AUTHKICK"
	case N_AUTHCHAL:
		return "N_AUTHCHAL"
	case N_AUTHANS:
		return "N_AUTHANS"
	case N_REQAUTH:
		return "N_REQAUTH"
	case N_PAUSEGAME:
		return "N_PAUSEGAME"
	case N_GAMESPEED:
		return "N_GAMESPEED"
	case N_ADDBOT:
		return "N_ADDBOT"
	case N_DELBOT:
		return "N_DELBOT"
	case N_INITAI:
		return "N_INITAI"
	case N_FROMAI:
		return "N_FROMAI"
	case N_BOTLIMIT:
		return "N_BOTLIMIT"
	case N_BOTBALANCE:
		return "N_BOTBALANCE"
	case N_MAPCRC:
		return "N_MAPCRC"
	case N_CHECKMAPS:
		return "N_CHECKMAPS"
	case N_SWITCHNAME:
		return "N_SWITCHNAME"
	case N_SWITCHMODEL:
		return "N_SWITCHMODEL"
	case N_SWITCHTEAM:
		return "N_SWITCHTEAM"
	case N_INITTOKENS:
		return "N_INITTOKENS"
	case N_TAKETOKEN:
		return "N_TAKETOKEN"
	case N_EXPIRETOKENS:
		return "N_EXPIRETOKENS"
	case N_DROPTOKENS:
		return "N_DROPTOKENS"
	case N_DEPOSITTOKENS:
		return "N_DEPOSITTOKENS"
	case N_STEALTOKENS:
		return "N_STEALTOKENS"
	case N_SERVCMD:
		return "N_SERVCMD"
	case N_DEMOPACKET:
		return "N_DEMOPACKET"
	case NUMMSG:
		return "NUMMSG"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

func IsConnectingMessage(code MessageCode) bool {
	for _, comparison := range []MessageCode{N_CONNECT, N_AUTHANS, N_PING} {
		if code == comparison {
			return true
		}
	}

	return false
}

func IsSpammyMessage(code MessageCode) bool {
	for _, comparison := range []MessageCode{N_PING, N_PONG, N_CLIENTPING, N_POS} {
		if code == comparison {
			return true
		}
	}

	return false
}

var SERVER_ONLY = []MessageCode{
	N_SERVINFO,
	N_INITCLIENT,
	N_WELCOME,
	N_MAPCHANGE,
	N_SERVMSG,
	N_DAMAGE,
	N_HITPUSH,
	N_SHOTFX,
	N_EXPLODEFX,
	N_DIED,
	N_SPAWNSTATE,
	N_FORCEDEATH,
	N_TEAMINFO,
	N_ITEMSPAWN,
	N_ITEMACC,
	N_TIMEUP,
	N_CDIS,
	N_CURRENTMASTER,
	N_PONG,
	N_RESUME,
	N_BASESCORE,
	N_BASEINFO,
	N_BASEREGEN,
	N_ANNOUNCE,
	N_SENDDEMOLIST,
	N_SENDDEMO,
	N_DEMOPLAYBACK,
	N_SENDMAP,
	N_DROPFLAG,
	N_SCOREFLAG,
	N_RETURNFLAG,
	N_RESETFLAG,
	N_INVISFLAG,
	N_CLIENT,
	N_AUTHCHAL,
	N_INITAI,
	N_EXPIRETOKENS,
	N_DROPTOKENS,
	N_STEALTOKENS,
	N_DEMOPACKET,
}

func IsServerOnly(code MessageCode) bool {
	for _, comparison := range SERVER_ONLY {
		if code == comparison {
			return true
		}
	}

	return false
}
