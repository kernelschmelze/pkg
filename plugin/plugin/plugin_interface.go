package plugin

type PluginInterface interface {
	Start() error
	Stop() error
	IsActivated() bool
	Configure(v interface{})
	Do(v interface{}) error
	DoAction(action string, v interface{}) error
}

type Plugin struct{}

func (p *Plugin) Start() error                       { return nil }
func (p *Plugin) Stop() error                        { return nil }
func (p *Plugin) IsActivated() bool                  { return false }
func (p *Plugin) Configure(interface{})              {}
func (p *Plugin) Do(interface{}) error               { return nil }
func (p *Plugin) DoAction(string, interface{}) error { return nil }
