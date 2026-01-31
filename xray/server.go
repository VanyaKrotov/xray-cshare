package xray

import (
	"runtime"
	"runtime/debug"
	"strings"

	code_errors "github.com/VanyaKrotov/xray_cshare/errors"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
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
	config, err := serial.DecodeJSONConfig(strings.NewReader(jsonConfig))
	if err != nil {
		return nil, code_errors.New(err.Error(), JsonParseError)
	}

	coreCfg, err := config.Build()
	if err != nil {
		return nil, code_errors.New(err.Error(), LoadConfigError)
	}

	instance, err := core.New(coreCfg)
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
