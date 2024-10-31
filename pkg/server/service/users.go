package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/cfoust/sour/pkg/game"
	"github.com/cfoust/sour/pkg/game/io"
	P "github.com/cfoust/sour/pkg/game/protocol"
	"github.com/cfoust/sour/pkg/gameserver"
	"github.com/cfoust/sour/pkg/utils"

	"github.com/cfoust/sour/pkg/config"
	"github.com/cfoust/sour/pkg/server/ingress"
	"github.com/cfoust/sour/pkg/server/servers"
	"github.com/cfoust/sour/pkg/server/verse"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sasha-s/go-deadlock"
)

type TrackedPacket struct {
	Packet P.Packet
	Error  chan error
}

type ConnectionEvent struct {
	Server *servers.GameServer
}

// The status of the user's connection to their game server.
type UserStatus uint8

const (
	UserStatusConnecting = iota
	UserStatusConnected
	UserStatusDisconnected
)

type User struct {
	*utils.Session

	sessionID string
	Id        ingress.ClientID
	// Whether the user is connected (or connecting) to a game server
	Status UserStatus
	Name   string

	Connection ingress.Connection

	// Created when the user connects to a server and canceled when they
	// leave, regardless of reason (network or being disconnected by the
	// server)
	// This is NOT the same thing as Client.Connection.SessionContext(), which refers to
	// the lifecycle of the client's ingress connection
	ServerSession utils.Session
	Server        *servers.GameServer
	ServerClient  *gameserver.Client

	ELO *ELOState

	// True when the user is loading the map
	delayMessages bool
	messageQueue  []string

	serverConnections chan ConnectionEvent

	to chan TrackedPacket

	// The last server description sent to the user
	lastInfo    *P.ServerInfo
	sendingMap  bool
	autoexecKey string

	Space *verse.SpaceInstance

	From *P.MessageProxy
	To   *P.MessageProxy

	RawFrom *utils.Topic[io.RawPacket]
	RawTo   *utils.Topic[io.RawPacket]

	Mutex      deadlock.RWMutex
	queueMutex deadlock.RWMutex
	o          *UserOrchestrator
}

func (c *User) ReceiveConnections() <-chan ConnectionEvent {
	return c.serverConnections
}

func (u *User) GetSessionID() string {
	return u.sessionID[:5]
}

func (u *User) Logger() zerolog.Logger {
	u.Mutex.RLock()
	logger := log.With().
		Str("session", u.GetSessionID()).
		Str("type", u.Connection.DeviceType()).
		Str("name", u.Name).
		Logger()

	if u.Server != nil {
		logger = logger.With().Str("server", u.Server.Reference()).Logger()
	}
	u.Mutex.RUnlock()

	return logger
}

func (u *User) GetClientNum() int {
	u.Mutex.RLock()
	num := int(u.ServerClient.CN)
	u.Mutex.RUnlock()
	return num
}

func (u *User) GetServerInfo() P.ServerInfo {
	u.Mutex.RLock()
	info := u.lastInfo
	u.Mutex.RUnlock()
	return *info
}

func (c *User) GetStatus() UserStatus {
	c.Mutex.RLock()
	status := c.Status
	c.Mutex.RUnlock()
	return status
}

func (c *User) DelayMessages() {
	c.Mutex.Lock()
	c.delayMessages = true
	c.Mutex.Unlock()
}

func (c *User) RestoreMessages() {
	c.Mutex.Lock()
	c.delayMessages = false
	c.Mutex.Unlock()
	c.sendQueuedMessages()
}

func (c *User) SendChannel(channel uint8, messages ...P.Message) <-chan error {
	out := make(chan error, 1)
	c.to <- TrackedPacket{
		Packet: P.Packet{
			Channel:  channel,
			Messages: messages,
		},
		Error: out,
	}
	return out
}

func (c *User) SendChannelSync(channel uint8, messages ...P.Message) error {
	return <-c.SendChannel(channel, messages...)
}

func (c *User) Send(messages ...P.Message) <-chan error {
	return c.SendChannel(1, messages...)
}

func (c *User) SendSync(messages ...P.Message) error {
	return c.SendChannelSync(1, messages...)
}

// ResponseTimeout sends a message for a user and waits for a response of type `code`.
func (u *User) ResponseTimeout(ctx context.Context, timeout time.Duration, code P.MessageCode, messages ...P.Message) (P.Message, error) {
	errorChan := make(chan error)
	msgChan := make(chan P.Message)
	go func() {
		msg, err := u.From.NextTimeout(ctx, timeout, code)
		if err != nil {
			errorChan <- err
			return
		}
		msgChan <- msg
	}()
	go func() {
		err := <-u.Send(messages...)
		if err != nil {
			errorChan <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errorChan:
		return nil, err
	case msg := <-msgChan:
		return msg, nil
	}
}

func (u *User) Response(ctx context.Context, code P.MessageCode, messages ...P.Message) (P.Message, error) {
	return u.ResponseTimeout(ctx, 5*time.Second, code, messages...)
}

func (c *User) ReceiveToMessages() <-chan TrackedPacket {
	return c.to
}

func (u *User) IsAtHome(ctx context.Context) (bool, error) {
	return false, nil
}

func (u *User) GetServer() *servers.GameServer {
	u.Mutex.RLock()
	server := u.Server
	u.Mutex.RUnlock()
	return server
}

func (u *User) ServerSessionContext() context.Context {
	u.Mutex.RLock()
	ctx := u.ServerSession.Ctx()
	u.Mutex.RUnlock()
	return ctx
}

func (u *User) GetSpace() *verse.SpaceInstance {
	u.Mutex.RLock()
	space := u.Space
	u.Mutex.RUnlock()
	return space
}

// SPAAAAAAAAACE
func (u *User) IsInSpace() bool {
	return u.GetSpace() != nil
}

func (u *User) GetServerName() string {
	serverName := "???"

	space := u.GetSpace()
	if space != nil {
		return space.GetID()
	}

	server := u.GetServer()
	if server != nil {
		serverName = server.GetFormattedReference()
	} else {
		if u.Connection.Type() == ingress.ClientTypeWS {
			serverName = "web"
		}
	}

	return serverName
}

func (u *User) GetFormattedName() string {
	name := u.GetName()
	return name
}

func (c *User) sendQueuedMessages() {
	c.queueMutex.Lock()
	for _, message := range c.messageQueue {
		c.sendMessage(message)
	}
	c.messageQueue = make([]string, 0)
	c.queueMutex.Unlock()
}

func (c *User) sendMessage(message string) {
	c.Send(P.ServerMessage{Text: message})
}

func (u *User) queueMessage(message string) {
	u.Mutex.RLock()
	delayed := u.delayMessages
	u.Mutex.RUnlock()

	if delayed {
		u.queueMutex.Lock()
		u.messageQueue = append(u.messageQueue, message)
		u.queueMutex.Unlock()
		return
	}

	u.sendMessage(message)
}

func (u *User) Message(message string) {
	u.queueMessage(fmt.Sprintf("%s %s", game.Magenta("~>"), message))
}

func (u *User) RawMessage(message string) {
	u.queueMessage(message)
}

func (u *User) Reference() string {
	return fmt.Sprintf("%s (%s)", u.GetName(), u.GetServerName())
}

func (u *User) GetFormattedReference() string {
	return fmt.Sprintf("%s (%s)", u.GetFormattedName(), u.GetServerName())
}

func (u *User) GetName() string {
	u.Mutex.RLock()
	name := u.Name
	u.Mutex.RUnlock()
	return name
}

func (u *User) SetName(ctx context.Context, name string) error {
	return fmt.Errorf("feature disabled")
}

func (u *User) AnnounceELO() {
	u.Mutex.RLock()
	result := "ratings: "
	for _, duel := range u.o.Duels {
		name := duel.Name
		state := u.ELO.Ratings[name]
		result += fmt.Sprintf(
			"%s %d (%s-%s-%s) ",
			name,
			state.Rating,
			game.Green(fmt.Sprint(state.Wins)),
			game.Yellow(fmt.Sprint(state.Draws)),
			game.Red(fmt.Sprint(state.Losses)),
		)
	}
	u.Mutex.RUnlock()

	u.Message(result)
}

func (u *User) ConnectToSpace(server *servers.GameServer, id string) (<-chan bool, error) {
	return u.ConnectToServer(server, id, false, true)
}

func (u *User) Connect(server *servers.GameServer) (<-chan bool, error) {
	return u.ConnectToServer(server, "", false, false)
}

func (u *User) ConnectToServer(server *servers.GameServer, target string, shouldCopy bool, isSpace bool) (<-chan bool, error) {
	if u.Connection.NetworkStatus() == ingress.NetworkStatusDisconnected {
		log.Warn().Msgf("client not connected to cluster but attempted connect")
		return nil, fmt.Errorf("client not connected to cluster")
	}

	u.DelayMessages()

	oldServer := u.GetServer()
	if oldServer != nil {
		oldServer.Leave(uint32(u.Id))
		u.ServerSession.Cancel()

		// Remove all the other clients from this client's perspective
		u.o.Mutex.Lock()
		users, ok := u.o.Servers[oldServer]
		if ok {
			newUsers := make([]*User, 0)
			for _, otherUser := range users {
				if u == otherUser {
					continue
				}

				u.Send(
					P.ClientDisconnected{
						Client: int32(otherUser.GetClientNum()),
					},
				)
				newUsers = append(newUsers, otherUser)
			}
			u.o.Servers[u.Server] = newUsers
		}
		u.o.Mutex.Unlock()
	}

	u.Mutex.Lock()
	u.Space = nil
	u.Server = server
	u.Status = UserStatusConnecting
	u.ServerSession = utils.NewSession(u.Session.Ctx())
	u.Mutex.Unlock()

	connected := make(chan bool, 1)

	serverClient, serverConnected := server.Connect(uint32(u.Id))
	u.ServerClient = serverClient

	serverName := server.Reference()
	if target != "" {
		serverName = target
	}
	u.Connection.Connect(serverName, server.Hidden, shouldCopy)

	// Give the client one second to connect.
	go func() {
		connectCtx, cancel := context.WithTimeout(u.ServerSession.Ctx(), time.Second*1)
		defer cancel()

		select {
		case <-serverConnected:
			u.Mutex.Lock()
			u.Status = UserStatusConnected
			u.Mutex.Unlock()

			u.o.Mutex.Lock()
			users, ok := u.o.Servers[server]
			newUsers := make([]*User, 0)
			if ok {
				for _, otherUser := range users {
					if u == otherUser {
						continue
					}

					newUsers = append(newUsers, otherUser)
				}
			}
			newUsers = append(newUsers, u)
			u.o.Servers[u.Server] = newUsers
			u.o.Mutex.Unlock()

			connected <- true
			u.serverConnections <- ConnectionEvent{
				Server: server,
			}

		case <-u.Session.Ctx().Done():
			connected <- false
		case <-connectCtx.Done():
			u.RestoreMessages()
			connected <- false
		}
	}()

	return connected, nil
}

// Mark the client's status as disconnected and cancel its session context.
// Called both when the client disconnects from ingress AND when the server kicks them out.
func (u *User) DisconnectFromServer() error {
	server := u.GetServer()
	if server != nil {
		server.Leave(uint32(u.Id))
	}

	u.Mutex.Lock()
	u.Server = nil
	u.Space = nil
	u.Status = UserStatusDisconnected
	u.Mutex.Unlock()

	u.ServerSession.Cancel()

	return nil
}

type UserOrchestrator struct {
	Duels   []config.DuelType
	Users   []*User
	Servers map[*servers.GameServer][]*User
	Mutex   deadlock.RWMutex
}

func NewUserOrchestrator(duels []config.DuelType) *UserOrchestrator {
	return &UserOrchestrator{
		Duels:   duels,
		Users:   make([]*User, 0),
		Servers: make(map[*servers.GameServer][]*User),
	}
}

func (u *UserOrchestrator) PollUser(ctx context.Context, user *User) {
	<-user.Ctx().Done()
	u.RemoveUser(user)
}

func (u *UserOrchestrator) newSessionID() (ingress.ClientID, error) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()

	for attempts := 0; attempts < math.MaxUint16; attempts++ {
		number, _ := rand.Int(rand.Reader, big.NewInt(math.MaxUint16))
		truncated := ingress.ClientID(number.Uint64())

		taken := false
		for _, user := range u.Users {
			if user.Id == truncated {
				taken = true
			}
		}
		if taken {
			continue
		}

		return truncated, nil
	}

	return 0, fmt.Errorf("Failed to assign client ID")
}

func (u *UserOrchestrator) AddUser(ctx context.Context, connection ingress.Connection) (*User, error) {
	id, err := u.newSessionID()
	if err != nil {
		return nil, err
	}

	sessionID := utils.HashString(fmt.Sprintf("%d-%s", id, connection.Host()))

	u.Mutex.Lock()
	user := User{
		Id:                id,
		sessionID:         sessionID,
		Status:            UserStatusDisconnected,
		Connection:        connection,
		Session:           connection.Session(),
		ELO:               NewELOState(u.Duels),
		Name:              "unnamed",
		From:              P.NewMessageProxy(true),
		To:                P.NewMessageProxy(false),
		to:                make(chan TrackedPacket, 1000),
		serverConnections: make(chan ConnectionEvent),
		ServerSession:     utils.NewSession(ctx),
		o:                 u,
		RawFrom:           utils.NewTopic[io.RawPacket](),
		RawTo:             utils.NewTopic[io.RawPacket](),
	}
	u.Users = append(u.Users, &user)
	u.Mutex.Unlock()

	go u.PollUser(ctx, &user)

	logger := user.Logger()
	logger.Info().Msg("user joined")

	return &user, nil
}

func (u *UserOrchestrator) RemoveUser(user *User) {
	u.Mutex.Lock()

	newUsers := make([]*User, 0)
	for _, other := range u.Users {
		if other == user {
			continue
		}
		newUsers = append(newUsers, other)
	}
	u.Users = newUsers

	for server, users := range u.Servers {
		serverUsers := make([]*User, 0)
		for _, other := range users {
			if other == user {
				continue
			}
			serverUsers = append(serverUsers, other)
		}

		u.Servers[server] = serverUsers
	}

	u.Mutex.Unlock()
}

func (u *UserOrchestrator) FindUser(id ingress.ClientID) *User {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	for _, user := range u.Users {
		if user.Id != id {
			continue
		}

		return user
	}

	return nil
}
