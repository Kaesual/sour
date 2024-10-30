package verse

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cfoust/sour/cmd/server/config"
	"github.com/cfoust/sour/cmd/server/ingress"
	gameServers "github.com/cfoust/sour/cmd/server/servers"
	"github.com/cfoust/sour/pkg/assets"
	"github.com/cfoust/sour/pkg/game/commands"
	C "github.com/cfoust/sour/pkg/game/constants"
	"github.com/cfoust/sour/pkg/server"
	"github.com/cfoust/sour/pkg/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sasha-s/go-deadlock"
)

type SpaceInstance struct {
	utils.Session

	SpaceConfig

	id          string
	Space       *UserSpace
	PresetSpace *config.PresetSpace
	Editing     *EditingState
	Server      *gameServers.GameServer
}

func (s *SpaceInstance) IsOpenEdit() bool {
	if s.Editing == nil {
		return false
	}
	return s.Editing.IsOpenEdit()
}

func (s *SpaceInstance) GetID() string {
	return s.id
}

func (s *SpaceInstance) refreshConfig(ctx context.Context) error {
	if s.Space == nil {
		return nil
	}

	config, err := s.Space.GetConfig(ctx)
	if err != nil {
		return err
	}

	s.SpaceConfig = *config
	return nil
}

func (s *SpaceInstance) GetDescription(ctx context.Context) (string, error) {
	err := s.refreshConfig(ctx)
	if err != nil {
		return "", err
	}

	return s.Description, nil
}

func (s *SpaceInstance) GetAlias(ctx context.Context) (string, error) {
	err := s.refreshConfig(ctx)
	if err != nil {
		return "", err
	}

	return s.Alias, nil
}

// Combine the description and alias (or ID) to make the servinfo string.
func (s *SpaceInstance) GetServerInfo(ctx context.Context) (string, error) {
	alias, err := s.GetAlias(ctx)
	if err != nil {
		return "", err
	}

	description, err := s.GetDescription(ctx)
	if err != nil {
		return "", err
	}

	reference := s.id[:5]
	if alias != "" {
		reference = alias
		if len(reference) > 16 {
			reference = reference[:16]
		}
	}

	tail := fmt.Sprintf(" [%s]", reference)
	overshoot := (len(tail) + len(description)) - 25
	if overshoot > 0 {
		description = description[:len(description)-overshoot]
	}

	return description + tail, nil
}

func (s *SpaceInstance) GetMap(ctx context.Context) (string, error) {
	err := s.refreshConfig(ctx)
	if err != nil {
		return "", err
	}

	return s.Map, nil
}

func (s *SpaceInstance) GetLinks(ctx context.Context) ([]Link, error) {
	err := s.refreshConfig(ctx)
	if err != nil {
		return nil, err
	}

	return s.Links, nil
}

func (s *SpaceInstance) PollEdits(ctx context.Context) {
	edits := s.Server.Edits.Subscribe()
	for {
		select {
		case <-s.Ctx().Done():
			return
		case edit := <-edits.Recv():
			if s.Editing == nil {
				continue
			}
			s.Editing.Process(ingress.ClientID(edit.Client), edit.Message)
			continue
		}
	}
}

type SpaceManager struct {
	utils.Session

	// space id -> instance
	instances map[string]*SpaceInstance
	verse     *Verse
	servers   *gameServers.ServerManager
	mutex     deadlock.RWMutex
	maps      *assets.AssetFetcher
}

func NewSpaceManager(servers *gameServers.ServerManager, maps *assets.AssetFetcher) *SpaceManager {
	return &SpaceManager{
		Session:   utils.NewSession(context.Background()),
		servers:   servers,
		instances: make(map[string]*SpaceInstance),
		maps:      maps,
	}
}

func (s *SpaceManager) Logger() zerolog.Logger {
	return log.With().Str("service", "spaces").Logger()
}

func (s *SpaceManager) SearchSpace(ctx context.Context, id string) (*UserSpace, error) {
	// Search for a user's space matching this ID
	space, _ := s.verse.FindSpace(ctx, id)
	if space != nil {
		return space, nil
	}

	// We don't care if that errored, search the maps (which are implicitly spaces)
	found := s.maps.FindMap(id)
	if found == nil {
		return nil, fmt.Errorf("ambiguous reference")
	}

	// TODO support game maps
	return nil, fmt.Errorf("found map, but unsupported")
}

func (s *SpaceManager) FindInstance(server *gameServers.GameServer) *SpaceInstance {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, instance := range s.instances {
		if instance.Server == server {
			return instance
		}
	}

	return nil
}

func (s *SpaceManager) WatchInstance(ctx context.Context, space *SpaceInstance) {
	select {
	case <-ctx.Done():
		return
	case <-space.Ctx().Done():
		if space.Editing != nil {
			space.Editing.Checkpoint(ctx)
		}

		s.mutex.Lock()

		deleteId := ""
		for id, instance := range s.instances {
			if instance == space {
				deleteId = id
			}
		}

		if deleteId != "" {
			delete(s.instances, deleteId)
		}

		s.mutex.Unlock()
		return
	}
}

func (s *SpaceManager) StartSpace(ctx context.Context, id string) (*SpaceInstance, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.Logger()

	space, err := s.SearchSpace(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msgf("could not find space %s", id)
		return nil, err
	}

	if instance, ok := s.instances[space.GetID()]; ok {
		return instance, nil
	}

	config, err := space.GetConfig(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch config for space")
		return nil, err
	}

	verseMap, err := space.GetMap(ctx)
	if err != nil {
		return nil, err
	}

	map_, err := verseMap.LoadGameMap(ctx)
	if err != nil {
		return nil, err
	}

	editing := NewEditingState(s.verse, space, verseMap)
	err = editing.LoadMap(map_)
	if err != nil {
		return nil, err
	}

	instance := SpaceInstance{
		Session:     utils.NewSession(ctx),
		Space:       space,
		Editing:     editing,
		SpaceConfig: *config,
	}

	instance.id = space.GetID()

	description, err := instance.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}

	go editing.SavePeriodically(instance.Ctx())

	gameServer, err := s.servers.NewServer(instance.Ctx(), "", true)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create server for preset")
		return nil, err
	}

	gameServer.Alias = space.Reference()

	gameServer.SetDescription(description)
	// TODO gameServer.SendCommand("publicserver 1")
	gameServer.EmptyMap()

	instance.Server = gameServer

	go s.WatchInstance(ctx, &instance)

	go instance.PollEdits(instance.Ctx())

	s.instances[space.GetID()] = &instance

	return &instance, nil
}

func (s *SpaceManager) DoExploreMode(ctx context.Context, gameServer *gameServers.GameServer, skipRoot string) {
	maps := s.maps.GetMaps(skipRoot)

	skips := make(map[*server.Client]struct{})

	cycleMap := func() {
		var name string
		for {
			index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(maps))))
			map_ := maps[index.Int64()]

			gameServer.Mutex.RLock()
			currentMap := gameServer.Map
			gameServer.Mutex.RUnlock()

			name = map_.Name
			if name == "" || name == currentMap || strings.Contains(name, ".") || strings.Contains(name, " ") || map_.HasCFG {
				continue
			}

			break
		}

		gameServer.ChangeMap(C.MODE_COOP, name)
		skips = make(map[*server.Client]struct{})
	}

	err := gameServer.Commands.Register(
		commands.Command{
			Name:        "skip",
			Description: "vote to skip to the next map",
			Callback: func(client *server.Client) {
				if _, ok := skips[client]; ok {
					client.Message("you have already voted to skip")
					return
				}

				name := gameServer.Clients.UniqueName(client)
				gameServer.Message(fmt.Sprintf("%s voted to skip to the next map (say #skip to vote)", name))

				skips[client] = struct{}{}

				numClients := gameServer.Clients.GetNumClients()
				if len(skips) > numClients/2 || (numClients == 1 && len(skips) == 1) {
					cycleMap()
				}
			},
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("could not register explore command")
	}

	tick := time.NewTicker(3 * time.Minute)

	cycleMap()

	for {
		select {
		case <-gameServer.Session.Ctx().Done():
			return
		case <-tick.C:
			cycleMap()
			continue
		}
	}
}

func (s *SpaceManager) StartPresetSpace(ctx context.Context, presetSpace config.PresetSpace) (*SpaceInstance, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config := presetSpace.Config
	id := config.Alias

	links := make([]Link, 0)
	for _, link := range config.Links {
		links = append(links, Link{
			Teleport:    link.Teleport,
			Teledest:    link.Teledest,
			Destination: link.Destination,
		})
	}

	logger := s.Logger()

	gameServer, err := s.servers.NewServer(ctx, presetSpace.Preset, true)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create server for preset")
		return nil, err
	}

	gameServer.Alias = config.Alias

	if config.Description != "" {
		gameServer.SetDescription(config.Description)
	} else {
		gameServer.SetDescription(fmt.Sprintf("Sour [%s]", config.Alias))
	}

	logger.Info().Msgf("started space %s", config.Alias)

	if presetSpace.ExploreMode {
		go s.DoExploreMode(ctx, gameServer, presetSpace.ExploreModeSkip)
	}

	instance := SpaceInstance{
		Session:     utils.NewSession(s.Ctx()),
		Server:      gameServer,
		PresetSpace: &presetSpace,
		SpaceConfig: SpaceConfig{
			Alias:       config.Alias,
			Description: config.Description,
			Links:       links,
			Map:         "",
		},
	}

	go s.WatchInstance(ctx, &instance)

	instance.id = id
	s.instances[id] = &instance

	return &instance, nil
}
