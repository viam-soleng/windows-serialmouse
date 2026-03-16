package windowsserialmouse

import (
	"context"
	"fmt"

	generic "go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"golang.org/x/sys/windows/registry"
)

var (
	Disable = resource.NewModel("viam-soleng", "windows-serialmouse", "disable")
)

func init() {
	resource.RegisterComponent(generic.API, Disable,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newWindowsSerialmouseDisable,
		},
	)
}

type Config struct {
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	return nil, nil, nil
}

type windowsSerialmouseDisable struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()
}

func newWindowsSerialmouseDisable(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewDisable(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewDisable(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (resource.Resource, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &windowsSerialmouseDisable{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *windowsSerialmouseDisable) Name() resource.Name {
	return s.name
}

func (s *windowsSerialmouseDisable) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	// https://paulhutch.blog/2019/06/24/disable-serial-mouse-detection/
	// GPS will be on the serial line, not a serial mouse
	const keyPath = `System\CurrentControlSet\Services\sermouse`

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("Start")
	if err != nil {
		return nil, fmt.Errorf("failed to read Start value: %w", err)
	}

	// No change required
	if val == 4 {
		return map[string]interface{}{
			"start":   4,
			"changed": false,
			"message": "Start value previously set to 4",
		}, nil
	}

	if err := k.SetDWordValue("Start", 4); err != nil {
		return nil, fmt.Errorf("failed to set Start value: %w", err)
	}

	s.logger.Info("Windows serial mouse Start registry value changed from 3 to 4")
	return map[string]interface{}{
		"start":   4,
		"changed": true,
		"message": "Start value changed from 3 to 4",
	}, nil
}

func (s *windowsSerialmouseDisable) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
