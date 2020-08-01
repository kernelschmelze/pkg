package plugin

import (
	"reflect"
	"sort"
	"sync"
)

// PluginInfo wraps a priority level with a plugin interface.
type PluginInfo struct {
	Priority int
	Plugin   PluginInterface
}

// PluginList holds a statically-typed sorted map of plugins
// registered on Noise.
type PluginList struct {
	keys   map[reflect.Type]*PluginInfo
	values []*PluginInfo

	vGuard sync.RWMutex
	kGuard sync.RWMutex
}

// NewPluginList creates a new instance of a sorted plugin list.
func NewPluginList() *PluginList {
	return &PluginList{
		keys:   make(map[reflect.Type]*PluginInfo),
		values: make([]*PluginInfo, 0),
	}
}

// SortByPriority sorts the plugins list by each plugins priority.
func (m *PluginList) SortByPriority() {
	m.vGuard.Lock()
	sort.SliceStable(m.values, func(i, j int) bool {
		return m.values[i].Priority < m.values[j].Priority
	})
	m.vGuard.Unlock()
}

// PutInfo places a new plugins info onto the list.
func (m *PluginList) PutInfo(plugin *PluginInfo) bool {
	ty := reflect.TypeOf(plugin.Plugin)
	if _, ok := m.keys[ty]; ok {
		return false
	}

	m.kGuard.Lock()
	m.keys[ty] = plugin
	m.kGuard.Unlock()

	m.vGuard.Lock()
	m.values = append(m.values, plugin)
	m.vGuard.Unlock()

	return true
}

// Put places a new plugin with a set priority onto the list.
func (m *PluginList) Put(priority int, plugin PluginInterface) bool {
	return m.PutInfo(&PluginInfo{
		Priority: priority,
		Plugin:   plugin,
	})
}

// Len returns the number of plugins in the plugin list.
func (m *PluginList) Len() int {
	m.kGuard.RLock()
	length := len(m.keys)
	m.kGuard.RUnlock()
	return length
}

// GetInfo gets the priority and plugin interface given a plugin ID. Returns nil if not exists.
func (m *PluginList) GetInfo(withTy interface{}) (*PluginInfo, bool) {
	m.kGuard.RLock()
	item, ok := m.keys[reflect.TypeOf(withTy)]
	m.kGuard.RUnlock()
	return item, ok
}

// GetName returns the name of the plugin
func (m *PluginList) GetName(plugin PluginInterface) string {
	return reflect.TypeOf(plugin).String()
}

// Get returns the plugin interface given a plugin ID. Returns nil if not exists.
func (m *PluginList) Get(withTy interface{}) (PluginInterface, bool) {
	if info, ok := m.GetInfo(withTy); ok {
		return info.Plugin, true
	}
	return nil, false
}

// Each goes through every plugin in ascending order of priority of the plugin list.
func (m *PluginList) Each(f func(value PluginInterface)) {
	m.vGuard.RLock()
	for _, item := range m.values {
		f(item.Plugin)
	}
	m.vGuard.RUnlock()
}

// EachReverse goes reverse through every plugin in ascending order of priority of the plugin list
func (m *PluginList) EachReverse(f func(value PluginInterface)) {
	m.vGuard.RLock()
	for i := range m.values {
		i = len(m.values) - 1 - i
		f(m.values[i].Plugin)
	}
	m.vGuard.RUnlock()
}
