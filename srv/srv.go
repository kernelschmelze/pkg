package srv

import (
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrAlreadyExist = errors.New("already exist")
	ErrDoesNotExist = errors.New("does not exist")
	ErrInvalidAddr  = errors.New("invalid addr")
	ErrCertMissing  = errors.New("certificates missing")
)

type Config struct {
	Addr    string
	Handler http.Handler
	CrtFile string
	KeyFile string
}

type onListen func(addr string, crtFile string, keyFile string)
type onShutdown func(addr string, err error)

type Srv struct {
	handler map[string]*http.Server
	mu      sync.RWMutex
	wg      sync.WaitGroup

	cbOnListen   onListen
	cbOnShutdown onShutdown
}

func New(onListen onListen, onShutdown onShutdown) *Srv {
	return &Srv{
		handler:      make(map[string]*http.Server),
		cbOnListen:   onListen,
		cbOnShutdown: onShutdown,
	}
}

func (s *Srv) Add(config Config) error {

	addr := config.Addr

	if s.Exist(addr) {
		return ErrAlreadyExist
	}

	forceTLS := strings.HasPrefix(addr, "https://")

	if forceTLS && (len(config.CrtFile) == 0 || len(config.KeyFile) == 0) {
		return ErrCertMissing
	}

	if strings.HasPrefix(addr, "http://") {
		config.CrtFile, config.KeyFile = "", ""
	}

	if index := strings.Index(addr, "://"); index >= 0 && index+3 <= len(addr) {
		addr = addr[index+3:]
	} else {
		if _, err := strconv.ParseInt(add, 10, 0); err == nil {
			addr = ":" + addr
		}
	}

	if len(addr) == 0 {
		return ErrInvalidAddr
	}

	server := &http.Server{
		Addr: addr,
	}

	if certificate, err := tlsCertificate(config.CrtFile, config.KeyFile); err == nil && certificate != nil {
		// preparation for cert hot reload, todo
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			GetCertificate: func(ri *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return certificate, nil
			},
		}
	} else if forceTLS {
		if err == nil {
			err = ErrCertMissing
		}
		return err
	}

	if config.Handler != nil {
		server.Handler = config.Handler
	}

	s.mu.Lock()
	s.handler[addr] = server
	s.mu.Unlock()

	s.wg.Add(1)

	go func(server *http.Server, crtFile string, keyFile string) {

		defer s.wg.Done()

		addr := server.Addr

		if server.TLSConfig != nil {
			s.onListen(addr, crtFile, keyFile)
			err := server.ListenAndServeTLS("", "")
			s.onShutdown(addr, err)
		} else {
			s.onListen(addr, "", "")
			err := server.ListenAndServe()
			s.onShutdown(addr, err)
		}

	}(server, config.CrtFile, config.KeyFile)

	return nil
}

func (s *Srv) Remove(addr string) error {

	s.mu.RLock()
	server, exist := s.handler[addr]
	s.mu.RUnlock()

	if exist {
		s.shutdown(server)

		s.mu.Lock()
		delete(s.handler, addr)
		s.mu.Unlock()

		return nil
	}

	return ErrDoesNotExist
}

func (s *Srv) Exist(addr string) bool {

	s.mu.RLock()
	_, exist := s.handler[addr]
	s.mu.RUnlock()

	return exist
}

func (s *Srv) Close() {

	s.mu.RLock()
	for i := range s.handler {
		server := s.handler[i]
		s.shutdown(server)
	}
	s.mu.RUnlock()

	s.mu.Lock()
	s.handler = make(map[string]*http.Server)
	s.mu.Unlock()

	s.wg.Wait()
}

func (s *Srv) shutdown(server *http.Server) {

	if server == nil {
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	server.Shutdown(ctx)

}

func (s *Srv) onListen(addr string, crtFile string, keyFile string) {

	if s.cbOnListen != nil {
		s.cbOnListen(addr, crtFile, keyFile)
	}

}

func (s *Srv) onShutdown(addr string, err error) {

	if s.cbOnShutdown != nil {
		s.cbOnShutdown(addr, err)
	}

}

func tlsCertificate(crtFile string, keyFile string) (*tls.Certificate, error) {

	if len(crtFile) == 0 && len(keyFile) == 0 {
		return nil, nil
	}

	var err error

	if crtFile, err = expandPath(crtFile); err != nil {
		return nil, err
	}

	if keyFile, err = expandPath(keyFile); err != nil {
		return nil, err
	}

	crt, err := ioutil.ReadFile(crtFile)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(crt, key)

	return &cert, err

}

func expandPath(path string) (string, error) {

	if path == "" {
		return "", nil
	}

	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "~", usr.HomeDir, 1)
	}

	return filepath.Abs(path)

}
