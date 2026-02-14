package wrapsocket
 
import (
	"context"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/leoheung/go-patterns/utils"
)

type HeartbeatConfig struct {
	Interval  time.Duration
	Timeout   time.Duration
	MaxMissed int
}

func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Interval:  30 * time.Second,
		Timeout:   10 * time.Second,
		MaxMissed: 3,
	}
}

type Heartbeat struct {
	config   *HeartbeatConfig
	conn     *Conn
	missed   int
	lastPong time.Time
	stopChan chan struct{}
}

func NewHeartbeat(conn *Conn, config *HeartbeatConfig) *Heartbeat {
	if config == nil {
		config = DefaultHeartbeatConfig()
	}
	return &Heartbeat{
		config:   config,
		conn:     conn,
		missed:   0,
		lastPong: time.Now(),
		stopChan: make(chan struct{}),
	}
}

func (h *Heartbeat) Start(ctx context.Context) {
	ticker := time.NewTicker(h.config.Interval)
	defer ticker.Stop()

	utils.DevLogInfo(fmt.Sprintf("[Heartbeat] Started for connection: %s", h.conn.ID))

	for {
		select {
		case <-ctx.Done():
			utils.DevLogInfo(fmt.Sprintf("[Heartbeat] Context cancelled for: %s", h.conn.ID))
			return
		case <-h.stopChan:
			utils.DevLogInfo(fmt.Sprintf("[Heartbeat] Stopped for connection: %s", h.conn.ID))
			return
		case <-ticker.C:
			if err := h.ping(ctx); err != nil {
				h.missed++
				utils.DevLogError(fmt.Sprintf("[Heartbeat] Ping failed for %s (missed: %d/%d): %v",
					h.conn.ID, h.missed, h.config.MaxMissed, err))

				if h.missed >= h.config.MaxMissed {
					utils.DevLogError(fmt.Sprintf("[Heartbeat] Connection %s exceeded max missed pings, closing", h.conn.ID))
					_ = h.conn.Close(websocket.StatusPolicyViolation, "heartbeat timeout")
					return
				}
			} else {
				h.missed = 0
				h.lastPong = time.Now()
			}
		}
	}
}

func (h *Heartbeat) Stop() {
	close(h.stopChan)
}

func (h *Heartbeat) ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, h.config.Timeout)
	defer cancel()

	return h.conn.ws.Ping(ctx)
}

func (h *Heartbeat) LastPong() time.Time {
	return h.lastPong
}

func (h *Heartbeat) MissedCount() int {
	return h.missed
}
