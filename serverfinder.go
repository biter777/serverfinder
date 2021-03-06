// Serverfinder - find a network server (tcp/udp/http/etc)

/*********************** E X A M P L E ***********************\
func getStatus(addr string, port int, chkStatus string) (*Status, error) {
	req, err := http.NewRequest(http.MethodGet, "http://"+addr+":"+strconv.Itoa(port)+chkStatus, nil) // "GET"
	if err != nil {
		return nil, err
	}
	req.Close = true

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	return unmarshalResponse(resp)
}

func finderConfig() *serverfinder.Config {
	request := func(port int) error {
		_, err := getStatus("127.0.0.1", port, "/pr/v1/status")
		return err
	}
	return &serverfinder.Config{
		PortStart: 8900,
		PortEnd:   8900 + 10000,
		Request:   request,
	}
}

func main() {
	port, err := serverfinder.Find(finderConfig())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server port: %v", port)
}
\************************* E N J O Y ***********************/

package serverfinder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
)

// ErrNotFound - ErrNotFound
var ErrNotFound = fmt.Errorf("Server not found")

type finder struct {
	*Config
	respCh *responseChan
}

// ------------------------------------------------------------------

// Find - find a server (tcp/udp/http/etc), check server via Config.Request() func
func Find(cfg *Config) (port, proxyPort int, err error) {
	finder := newFinder(cfg)
	defer finder.close()
	return finder.find()
}

// ------------------------------------------------------------------

// newFinder - create a finder
func newFinder(cfg *Config) *finder {
	if err := cfg.chk(); err != nil {
		panic(err)
	}
	return &finder{
		Config: cfg,
	}
}

// ------------------------------------------------------------------

// close - close a finder
func (f *finder) close() {
	if f != nil && f.respCh != nil {
		f.respCh.close()
		runtime.GC()
	}
}

// ------------------------------------------------------------------

func (f *finder) find() (port , proxyPort int, err error) {
	// port, err = f.findViaEnv()
	// if err != nil {
	// 	port, err = f.findViaTempDir()
	// }
	// if err != nil {
	port, proxyPort, err = f.findViaNetwork()
	// }
	return port, proxyPort, err
}

// ------------------------------------------------------------------

// findViaEnv - search a monitorus server via environment settings
func (f *finder) findViaEnv() (port, proxyPort int, err error) {
	if f.EnvVar == "" {
		return -1, -1, ErrNotFound
	}
	port, err = strconv.Atoi(os.Getenv(f.EnvVar))
	if port < 1 {
		return -1, -1, ErrNotFound
	}
	proxyPort, err = f.Request(port)
	return port, proxyPort, err
}

// ------------------------------------------------------------------

func (f *finder) findViaTempDir() (port, proxyPort int, err error) {
	if f.EnvVar == "" {
		return -1, -1, ErrNotFound
	}

	tempDir := os.TempDir()
	data, err := ioutil.ReadFile(tempDir + string(os.PathSeparator) + f.EnvVar)
	if err != nil {
		return -1, -1, ErrNotFound
	}

	err = json.Unmarshal(data, &port)
	if err != nil {
		return -1, -1, ErrNotFound
	}

	proxyPort, err = f.Request(port)
	return port, proxyPort, err
}

// ------------------------------------------------------------------

// findViaNetwork - search a monitorus server via network
func (f *finder) findViaNetwork() (port , proxyPort int, err error) {
	if f.respCh == nil {
		f.respCh = newResponseChan(100)
	}

	var stop bool
	go func() {
		for port := f.PortStart; port < f.PortEnd && !stop; port++ {
			f.respCh.wait()
			go f.request(port)
		}
	}()

	for i := 0; i < f.PortEnd-f.PortStart; i++ {
		resp := f.respCh.rcv()
		if stop = resp.err == nil; stop {
			return resp.port, resp.portOptional, nil
		}
	}
	return -1, -1, ErrNotFound
}

// ------------------------------------------------------------------

func (f *finder) request(port int) {
	portOptional, err := f.Request(port)
	f.respCh.send(&response{err: err, port: port, portOptional: portOptional})
}

// ------------------------------------------------------------------
