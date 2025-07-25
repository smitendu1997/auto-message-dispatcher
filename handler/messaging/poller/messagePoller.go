package poller

import (
	"context"
	"sync"
	"time"

	"github.com/smitendu1997/auto-message-dispatcher/logger"
	messagingService "github.com/smitendu1997/auto-message-dispatcher/services/messaging"
)

// MessageHandler represents the Message polling worker
type MessageHandler struct {
	running          bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	mu               sync.RWMutex
	messagingService messagingService.MessagingSvcDriver
}

// NewMessageHandler creates a new Message Handler  instance
func NewMessageHandler(messagingService messagingService.MessagingSvcDriver) *MessageHandler {
	const functionName = "worker.NewMessageHandler"
	logger.Info(functionName, "creating_message_handler")

	worker := &MessageHandler{
		stopChan:         make(chan struct{}),
		messagingService: messagingService,
	}

	logger.Info(functionName, " created")
	return worker
}

// Start begins the Message polling process
func (w *MessageHandler) Start() error {
	const functionName = "worker.MessageHandler.Start"

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		logger.Info(functionName, "worker_already_running")
		return nil
	}

	logger.Info(functionName, "starting_message_poller")
	w.running = true
	w.stopChan = make(chan struct{})

	w.wg.Add(1)
	go w.pollAndSendMessages()

	logger.Info(functionName, "message_poller_started")
	return nil
}

// Stop stops the Message polling process
func (w *MessageHandler) Stop() {
	const functionName = "worker.MessageHandler.Stop"

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		logger.Info(functionName, "worker_not_running")
		return
	}

	logger.Info(functionName, "stopping_message_poller")
	w.running = false
	close(w.stopChan)
	w.wg.Wait()
	logger.Info(functionName, "message_poller_stopped")
}

// pollAndSendMessages continuously polls for messages and sends it
func (w *MessageHandler) pollAndSendMessages() {
	const functionName = "worker.MessageHandler.pollAndSendMessages"
	defer w.wg.Done()
	ctx := context.Background()

	logger.Info(functionName, "starting_message_polling")
	w.messagingService.PollAndProcessMessages(ctx)
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			logger.Info(functionName, "polling_stopped")
			return
		case <-ticker.C:
			w.messagingService.PollAndProcessMessages(ctx)
		}
	}
}

// IsRunning checks if the Message polling worker is running
func (w *MessageHandler) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}
