package plugin

type FrpsConfig struct {
	BindAddr       string       `json:"bind_addr"`
	BindPort       uint         `json:"bind_port"`
	VhostHTTPPort  int          `json:"vhost_http_port"`
	VhostHTTPSPort int          `json:"vhost_https_port"`
	HttpPlugins    []HttpPlugin `json:"http_plugins"`
}

type HttpPlugin struct {
	Name string   `json:"name"`
	Addr string   `json:"addr"`
	Path string   `json:"path"`
	Ops  []string `json:"ops"`
}
