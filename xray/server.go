package xray

import (
	"encoding/json"
	"log"
	"runtime"
	"runtime/debug"

	"github.com/VanyaKrotov/xray_cshare/transfer"

	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	"github.com/xtls/xray-core/core"
	iconf "github.com/xtls/xray-core/infra/conf"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/proxy/socks"
)

const (
	JsonParseError     int = 1
	LoadConfigError    int = 2
	InitXrayError      int = 3
	StartXrayError     int = 4
	XrayAlreadyStarted int = 5
)

func Start(jsonConfig string) (*core.Instance, *transfer.Response) {
	var cfg iconf.Config
	if err := json.Unmarshal([]byte(jsonConfig), &cfg); err != nil {
		return nil, transfer.New(JsonParseError, err.Error())
	}

	coreCfg, err := cfg.Build()
	if err != nil {
		return nil, transfer.New(LoadConfigError, err.Error())
	}

	instance, err := core.New(coreCfg)
	if err != nil {
		log.Printf("Xray init error: %v", err)

		return nil, transfer.New(InitXrayError, err.Error())
	}

	if err := instance.Start(); err != nil {
		instance.Close()
		log.Printf("Xray start error: %v", err)

		return nil, transfer.New(StartXrayError, err.Error())
	}

	return instance, transfer.Success("Server started")
}

func Stop(instance *core.Instance) {
	instance.Close()
	instance = nil

	runtime.GC()
	debug.FreeOSMemory()
}
