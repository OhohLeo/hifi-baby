package bluetooth

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"
)

// NotificationManager manages BLE notifications to subscribed clients
type NotificationManager struct {
	subscribers map[string][]*bluetooth.Characteristic
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(ctx context.Context) *NotificationManager {
	ctx, cancel := context.WithCancel(ctx)
	return &NotificationManager{
		subscribers: make(map[string][]*bluetooth.Characteristic),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Subscribe adds a characteristic to the notification list
func (nm *NotificationManager) Subscribe(charUUID string, char *bluetooth.Characteristic) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.subscribers[charUUID] = append(nm.subscribers[charUUID], char)
	log.Debug().
		Str("characteristic", charUUID).
		Int("subscribers", len(nm.subscribers[charUUID])).
		Msg("Client subscribed to notifications")
}

// Unsubscribe removes a characteristic from the notification list
func (nm *NotificationManager) Unsubscribe(charUUID string, char *bluetooth.Characteristic) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	subs := nm.subscribers[charUUID]
	for i, c := range subs {
		if c == char {
			nm.subscribers[charUUID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	log.Debug().
		Str("characteristic", charUUID).
		Int("subscribers", len(nm.subscribers[charUUID])).
		Msg("Client unsubscribed from notifications")
}

// Notify sends data to all subscribed clients for a characteristic
func (nm *NotificationManager) Notify(charUUID string, data []byte) {
	nm.mu.RLock()
	chars := nm.subscribers[charUUID]
	nm.mu.RUnlock()

	if len(chars) == 0 {
		return
	}

	for _, char := range chars {
		// Write notification to characteristic
		// Note: TinyGo Bluetooth handles notification sending internally
		// when characteristic value is updated
		_, err := char.Write(data)
		if err != nil {
			log.Error().
				Err(err).
				Str("characteristic", charUUID).
				Msg("Failed to send notification")
		}
	}

	log.Debug().
		Str("characteristic", charUUID).
		Int("subscribers", len(chars)).
		Int("data_size", len(data)).
		Msg("Notification sent")
}

// HasSubscribers checks if a characteristic has any subscribers
func (nm *NotificationManager) HasSubscribers(charUUID string) bool {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	return len(nm.subscribers[charUUID]) > 0
}

// StartPeriodicNotifications starts sending periodic notifications for a characteristic
func (nm *NotificationManager) StartPeriodicNotifications(
	charUUID string,
	getData func() []byte,
	interval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-nm.ctx.Done():
				return
			case <-ticker.C:
				if nm.HasSubscribers(charUUID) {
					data := getData()
					nm.Notify(charUUID, data)
				}
			}
		}
	}()

	log.Info().
		Str("characteristic", charUUID).
		Dur("interval", interval).
		Msg("Started periodic notifications")
}

// Close shuts down the notification manager
func (nm *NotificationManager) Close() {
	nm.cancel()
	log.Info().Msg("Notification manager closed")
}
