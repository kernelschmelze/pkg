package plugin

import (
	"sync"

	"github.com/kernelschmelze/pkg/atom"
	"github.com/kernelschmelze/pkg/plugin/plugin"

	"github.com/pkg/errors"
)

var (
	manager *Manager
)

type Manager struct {
	activated   atom.Bool
	plugins     *plugin.PluginList
	pluginCount int
	jobs        chan Message
	wg          sync.WaitGroup
	kill        chan bool
}

func RegisterPlugin(plg interface{}, priority int) error {

	if manager == nil {
		manager = NewManager()
	}

	p, ok := plg.(plugin.PluginInterface)
	if !ok {
		return errors.Errorf("plugin '%T' is not a plugin.PluginInterface", plg)
	}

	var err error
	if priority == 0 {
		err = manager.AddPlugin(p)
	} else {
		err = manager.AddPluginWithPriority(priority, p)
	}

	return err
}

func GetManager() *Manager {

	if manager == nil {
		manager = NewManager()
	}

	return manager
}

func NewManager() *Manager {
	return &Manager{
		plugins: plugin.NewPluginList(),
		jobs:    make(chan Message, 128),
	}
}

func Dispatch(v interface{}) {
	manager := GetManager()
	manager.Dispatch(v)
}

func (m *Manager) Start() {

	m.kill = make(chan bool)

	m.wg.Add(1)
	m.activated.Set(true)

	go m.dispatcher()

	m.startPlugins()
}

func (m *Manager) Stop() {

	close(m.kill)

	m.activated.Set(false)
	m.wg.Wait()

	m.stopPlugins()
}

func (m *Manager) AddPlugin(plugin plugin.PluginInterface) error {

	err := m.AddPluginWithPriority(m.pluginCount, plugin)
	if err == nil {
		m.pluginCount++
	}

	return err
}

func (m *Manager) AddPluginWithPriority(priority int, plugin plugin.PluginInterface) error {

	if !m.plugins.Put(priority, plugin) {
		return errors.Errorf("plugin %s is already registered", m.plugins.GetName(plugin))
	}

	return nil
}

func (m *Manager) GetPlugin(plugin interface{}) (plugin.PluginInterface, bool) {

	p, exist := m.plugins.Get(plugin)
	return p, exist
}

func (m *Manager) ConfigurePlugin(plugin interface{}, config interface{}) error {
	if p, exist := m.plugins.Get(plugin); exist {
		p.Configure(config)
		return nil
	}
	return errors.Errorf("plugin '%T' not found", plugin)
}

func (m *Manager) startPlugins() {

	m.plugins.SortByPriority()
	m.plugins.Each(func(plugin plugin.PluginInterface) {
		if !plugin.IsActivated() {
			plugin.Start()
		}
	})
}

func (m *Manager) stopPlugins() {

	m.plugins.EachReverse(func(plugin plugin.PluginInterface) {
		if plugin.IsActivated() {
			plugin.Stop()
		}
	})
}
