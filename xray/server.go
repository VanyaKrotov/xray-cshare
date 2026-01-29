package xray

import (
	"runtime"
	"runtime/debug"

	code_errors "github.com/VanyaKrotov/xray_cshare/errors"

	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	"github.com/xtls/xray-core/core"
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

// Start xray server from json config
func Start(jsonConfig string) (*core.Instance, *code_errors.CodeError) {
	instance, err := core.StartInstance("json", []byte(jsonConfig))
	if err != nil {
		return nil, code_errors.New(err.Error(), InitXrayError)
	}

	if err := instance.Start(); err != nil {
		instance.Close()

		return nil, code_errors.New(err.Error(), StartXrayError)
	}

	debug.FreeOSMemory()

	return instance, nil
}

func Stop(instance *core.Instance) {
	instance.Close()
	instance = nil

	runtime.GC()
	debug.FreeOSMemory()
}

/*
	var cfg iconf.Config
	if err := json.Unmarshal([]byte(jsonConfig), &cfg); err != nil {
		return nil, code_errors.New(err.Error(), JsonParseError)
	}

	coreCfg, err := cfg.Build()
	if err != nil {
		return nil, code_errors.New(err.Error(), LoadConfigError)
	}

	instance, err := core.New(coreCfg)
*/
