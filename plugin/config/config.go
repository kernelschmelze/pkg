package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/kernelschmelze/pkg/path"
	manager "github.com/kernelschmelze/pkg/plugin/manager"
	"github.com/kernelschmelze/pkg/plugin/watcher"

	"golang.org/x/crypto/blake2b"

	"github.com/pelletier/go-toml"
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
	writeGuard sync.RWMutex
	mu         sync.RWMutex
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

	tml, err := toml.LoadFile(path)

	if err != nil {
		return err
	}

	update := make(map[interface{}]interface{})

	c.ownerGuard.RLock()

	for plugin, config := range c.owner {

		name := getName(plugin)

		data := tml.Get(name)

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

	c.mu.Lock()
	c.fileName = path
	c.mu.Unlock()

	for plugin, config := range update {
		manager.GetManager().ConfigurePlugin(plugin, config)
	}

	return err

}

func (c *Config) Write(plugin interface{}, key string, value interface{}) error {

	c.writeGuard.Lock()
	defer c.writeGuard.Unlock()

	c.mu.RLock()
	configFile := c.fileName
	c.mu.RUnlock()

	tml, err := toml.LoadFile(configFile)

	if err != nil {
		return err
	}

	var path []string

	name := getName(plugin)
	key = strings.ToLower(key)
	keys := strings.Split(key, ".")

	path = append(path, name)
	path = append(path, keys...)

	tml.SetPath(path, value)

	data, err := tml.Marshal()
	if err != nil {
		return err
	}

	tmp := configFile + ".tmp"
	if err = ioutil.WriteFile(tmp, data, 0644); err == nil {
		copyFile(c.fileName, configFile+".old")
		err = os.Rename(tmp, configFile)
	}

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

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
