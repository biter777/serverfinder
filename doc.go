/*
Package serverfinder - find a network server (tcp/udp/http/etc)

Usage

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

Contributing

 Welcome pull requests, bug fixes and issue reports.
 Before proposing a change, please discuss it first by raising an issue.
*/
package serverfinder
