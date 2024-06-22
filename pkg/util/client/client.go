package client

import "github.com/denisbrodbeck/machineid"

var (
	_clientId = ""
	AppId     = "6e3455881913579499efaf1208f834a5" //md5("github.com/frp-client")

)

func ClientId() string {
	if _clientId == "" {
		id, _ := machineid.ProtectedID(AppId)
		_clientId = id
	}
	return _clientId
}
