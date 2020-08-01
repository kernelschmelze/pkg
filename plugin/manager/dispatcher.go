package plugin

import (
	"github.com/kernelschmelze/pkg/plugin/plugin"
)

type Message struct {
	Action  string
	Payload interface{}
}

func NewMessage(action string, v interface{}) Message {
	msg := Message{Action: action}
	msg.Payload = v
	return msg
}

func (m *Manager) Dispatch(v interface{}) {

	if !m.activated.IsSet() {
		return
	}

	switch v.(type) {

	case Message:
		m.jobs <- v.(Message)

	default:

		msg := Message{}
		msg.Payload = v

		m.jobs <- msg

	}
}

func (m *Manager) Do(v interface{}) error {

	var err error

	m.plugins.Each(func(plugin plugin.PluginInterface) {
		if err == nil && plugin.IsActivated() {
			err = plugin.Do(v)
		}
	})

	return err
}

func (m *Manager) DoAction(action string, v interface{}) error {

	var err error

	m.plugins.Each(func(plugin plugin.PluginInterface) {
		if err == nil {
			err = plugin.DoAction(action, v)
		}
	})

	return err
}

func (m *Manager) dispatcher() {

	defer m.wg.Done()

	for {

		select {

		case <-m.kill:
			return

		case msg := <-m.jobs:

			if len(msg.Action) == 0 {

				m.Do(msg.Payload)

			} else {

				m.DoAction(msg.Action, msg.Payload)

			}

		}

	}

}
