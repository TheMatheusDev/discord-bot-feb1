package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"feb-notify/internal/squadstats"
)

// Callback defines the function signature for sending notifications.
type Callback func(message string)

// Manager handles the global active map monitor.
type Manager struct {
	client      *squadstats.Client
	subscribers map[string]struct{}
	isRunning   bool
	cancel      context.CancelFunc
	mu          sync.Mutex
}

// NewManager creates a new monitor manager.
func NewManager(client *squadstats.Client) *Manager {
	return &Manager{
		client:      client,
		subscribers: make(map[string]struct{}),
	}
}

// StartMonitor begins a monitoring loop. If already running, adds the user and returns true.
// Returns (alreadyRunning, initialDetails, error)
func (m *Manager) StartMonitor(ctx context.Context, userID string, onMapChange Callback) (bool, *squadstats.MatchDetails, error) {
	m.mu.Lock()
	if m.isRunning {
		m.subscribers[userID] = struct{}{}
		m.mu.Unlock()
		return true, nil, nil
	}

	m.isRunning = true
	m.subscribers[userID] = struct{}{}
	monitorCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.mu.Unlock()

	// Initial fetch to get current map details
	initialDetails, err := m.fetchCurrentMapDetails(monitorCtx)
	if err != nil {
		m.StopMonitor()
		return false, nil, fmt.Errorf("falha ao obter dados iniciais: %w", err)
	}

	go m.runLoop(monitorCtx, initialDetails, onMapChange)
	return false, initialDetails, nil
}

// AddSubscriber adds a user to the active monitor.
func (m *Manager) AddSubscriber(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.isRunning {
		return fmt.Errorf("o monitoramento não está ativo no momento")
	}
	m.subscribers[userID] = struct{}{}
	return nil
}

// StopMonitor safely stops the global monitor and clears subscribers.
func (m *Manager) StopMonitor() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.isRunning = false
	m.subscribers = make(map[string]struct{})
}

func (m *Manager) fetchCurrentMapDetails(ctx context.Context) (*squadstats.MatchDetails, error) {
	servers, err := m.client.GetServerData(ctx)
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("nenhum dado de servidor retornado")
	}

	matchID := servers[0].MatchID
	if matchID <= 0 {
		return nil, fmt.Errorf("matchID inválido: %d", matchID)
	}

	matchResp, err := m.client.GetMatchDetails(ctx, matchID)
	if err != nil {
		return nil, err
	}

	return &matchResp.Match, nil
}

func (m *Manager) runLoop(ctx context.Context, initialDetails *squadstats.MatchDetails, onMapChange Callback) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer m.StopMonitor()

	currentMapClass := initialDetails.MapClassname
	currentLayerClass := initialDetails.LayerClassname
	friendlyMap := initialDetails.Map
	friendlyLayer := initialDetails.Layer

	slog.Info("Iniciando monitoramento global", "map", friendlyMap, "layer", friendlyLayer)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Monitoramento cancelado/encerrado")
			return
		case <-ticker.C:
			slog.Debug("Executando verificação periódica global")
			newDetails, err := m.fetchCurrentMapDetails(ctx)
			if err != nil {
				slog.Error("Erro ao buscar dados durante monitoramento", "err", err)
				continue
			}

			if newDetails.MapClassname != currentMapClass && newDetails.LayerClassname != currentLayerClass {
				slog.Info("Troca de mapa detectada", "newMap", newDetails.MapClassname)
				
				m.mu.Lock()
				mentions := ""
				for subID := range m.subscribers {
					mentions += fmt.Sprintf("<@%s> ", subID)
				}
				m.mu.Unlock()

				msg := fmt.Sprintf("%s\nO mapa foi trocado!\n**Anterior**: %s - %s\n**Novo**: %s - %s",
					mentions, friendlyMap, friendlyLayer, newDetails.Map, newDetails.Layer)
				
				onMapChange(msg)
				return // End loop after notifying once
			}
		}
	}
}
