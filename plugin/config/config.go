package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/kernelschmelze/pkg/path"
	"github.com/kernelschmelze/pkg/plugin/watcher"

	manager "github.com/kernelschmelze/pkg/plugin/manager"

	"github.com/pelletier/go-toml"
	"golang.org/x/crypto/blake2b"
)

var (
	config *Config
)

func init() {
	config = NewConfig()
}

type Config struct {
	fileName   string
	owner      map[interface{}]interface{}
	ownerGuard sync.RWMutex
	hash       map[string][]byte
	data       *toml.Tree
}

func NewConfig() *Config {
	return &Config{
		owner: make(map[interface{}]interface{}),
		hash:  make(map[string][]byte),
	}
}

func GetConfig() *Config {

	if config == nil {
		config = NewConfig()
	}

	return config
}

func Read(path string) error {

	var err error

	path, err = utils.ExpandPath(path)
	if err != nil {
		return err
	}

	config := GetConfig()
	err = config.Read(path)

	wErr := watcher.Add(path, func(file string) {
		config.Read(file)
	})

	if err == nil {
		err = wErr
	}

	return err
}

func Write(plugin interface{}, key string, value interface{}) error {

	config := GetConfig()

	err := config.Write(plugin, key, value)
	return err
}

func Close() {
	watcher.Close()
}

func RegisterPlugin(plugin interface{}, v interface{}) {
	config := GetConfig()
	config.RegisterPlugin(plugin, v)
}

func (c *Config) Read(path string) error {

	var err error
	c.fileName = path
	c.data, err = toml.LoadFile(path)

	update := make(map[interface{}]interface{})

	c.ownerGuard.RLock()

	for plugin, config := range c.owner {

		name := getName(plugin)

		data := c.data.Get(name)

		if cfg, ok := data.(*toml.Tree); ok {

			hash := blake2b.Sum256([]byte(cfg.String()))

			if oldHash, exist := c.hash[name]; exist && bytes.Equal(oldHash[:], hash[:]) {
				continue
			}

			c.hash[name] = hash[:]

			cfg.Unmarshal(config)
			update[plugin] = config

		}

	}

	c.ownerGuard.RUnlock()

	for plugin, config := range update {
		manager.GetManager().ConfigurePlugin(plugin, config)
	}

	return err

}

func (c *Config) Write(plugin interface{}, key string, value interface{}) error {

	var path []string

	name := getName(plugin)
	key = strings.ToLower(key)
	keys := strings.Split(key, ".")

	path = append(path, name)
	path = append(path, keys...)

	c.data.SetPath(path, value)

	data, err := c.data.Marshal()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.fileName, data, 0644)

	return err
}

func (c *Config) RegisterPlugin(plugin interface{}, v interface{}) {
	c.ownerGuard.Lock()
	c.owner[plugin] = v
	c.ownerGuard.Unlock()
}

func getName(plugin interface{}) string {

	key := fmt.Sprintf("%T", plugin)
	key = strings.TrimLeft(key, "*")
	if offset := strings.Index(key, "."); offset > 0 {
		key = key[:offset]
	}

	return key
}
