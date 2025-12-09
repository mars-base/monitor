package utils

import (
	evbus "github.com/asaskevich/EventBus"
)

var (
	_bus       = evbus.New()
	_busServer *evbus.Server
	_busClient *evbus.Client
)

func Subscribe(topic string, fn any) error {
	return _bus.Subscribe(topic, fn)
}

func Unsubscribe(topic string, fn any) error {
	return _bus.Unsubscribe(topic, fn)
}

func Publish(topic string, data any) {
	_bus.Publish(topic, data)
}

func HasCallback(topic string) bool {
	return _bus.HasCallback(topic)
}
