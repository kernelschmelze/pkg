package plugin

import (
	"sync"

	"github.com/kernelschmelze/pkg/atom"
	"github.com/kernelschmelze/pkg/plugin/config"
	"github.com/kernelschmelze/pkg/plugin/manager"
)

type cbGeneric func() error
type cbDo func(v interface{}) error
type cbConfig func(v interface{})

func NewPlugin() *PluginBase {
	return NewPluginWithPriority(0)
}

func NewPluginWithPriority(priority int) *PluginBase {
	p := &PluginBase{
		priority: priority,
		action:   make(map[string]cbDo),
	}
	return p
}

type PluginConfig struct {
	Plugin      interface{}
	OnStart     cbGeneric
	OnStop      cbGeneric
	OnConfigure cbConfig
	OnDo        cbDo
	Config      interface{}
}

type PluginBase struct {
	priority    int
	action      map[string]cbDo
	guard       sync.RWMutex
	activated   atom.Bool
	onStart     cbGeneric
	onStop      cbGeneric
	onDo        cbDo
	onConfigure cbConfig
}

func (p *PluginBase) Init(callback PluginConfig) error {

	p.onStart = callback.OnStart
	p.onStop = callback.OnStop
	p.onConfigure = callback.OnConfigure
	p.onDo = callback.OnDo

	if err := plugin.RegisterPlugin(callback.Plugin, p.priority); err != nil {
		return err
	}

	if callback.Config != nil {
		config.RegisterPlugin(callback.Plugin, callback.Config)
	}

	return nil
}

func (p *PluginBase) RegisterAction(action string) {

	p.RegisterActionCallback(action, nil)

}

func (p *PluginBase) RegisterActionCallback(action string, callback cbDo) {

	if callback == nil {
		callback = p.Do
	}

	p.guard.Lock()
	p.action[action] = callback
	p.guard.Unlock()

}

func (p *PluginBase) Start() error {

	if p.onStart != nil {
		if err := p.onStart(); err != nil {
			return err
		}
	}

	p.activated.Set(true)
	return nil
}

func (p *PluginBase) Stop() error {

	p.activated.Set(false)

	if p.onStop != nil {
		if err := p.onStop(); err != nil {
			return err
		}
	}

	return nil
}

func (p *PluginBase) IsActivated() bool {
	return p.activated.Value()
}

func (p *PluginBase) Configure(v interface{}) {

	if p.onConfigure != nil {
		p.onConfigure(v)
	}

}

func (p *PluginBase) Do(v interface{}) error {

	if p.onDo != nil {
		return p.onDo(v)
	}

	return nil
}

func (p *PluginBase) DoAction(action string, v interface{}) error {

	var err error

	p.guard.RLock()
	callback, exist := p.action[action]
	p.guard.RUnlock()

	if exist && callback != nil {
		err = callback(v)
	}

	return err
}
