package windowsserialmouse

import (
	"context"
	"errors"
	"fmt"

	generic "go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"golang.org/x/sys/windows/registry"
)

var (
	Disable = resource.NewModel("viam-soleng", "windows-serialmouse", "disable")
)

const sermouseKeyPath = `System\CurrentControlSet\Services\sermouse`

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
	startChanged, err := s.disableSermouseService()
	if err != nil {
		return nil, err
	}

	// Also stop Windows from polling the built-in serial port(s) for a serial
	// mouse during enumeration, otherwise GPS data on the line can be
	// misdetected. PNP0501 is the ACPI ID for the built-in 16550 serial port.
	enumerationDisabledPorts, err := s.skipSerialPortEnumeration()
	if err != nil {
		return nil, err
	}

	message := "sermouse service was already disabled (Start was 4); serial port enumeration polling disabled"
	if startChanged {
		message = "sermouse service disabled (Start changed from 3 to 4); serial port enumeration polling disabled"
	}

	return map[string]interface{}{
		"start":                      fmt.Sprintf(`%s\Start : 4`, sermouseKeyPath),
		"sermouse-changed":           startChanged,
		"message":                    message,
		"enumeration_disabled_ports": enumerationDisabledPorts,
	}, nil
}

// disableSermouseService sets the sermouse service Start value to 4 (disabled).
// It reports whether the value was actually changed.
func (s *windowsSerialmouseDisable) disableSermouseService() (bool, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, sermouseKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("Start")
	if err != nil {
		return false, fmt.Errorf("failed to read Start value: %w", err)
	}

	// No change required
	if val == 4 {
		return false, nil
	}

	if err := k.SetDWordValue("Start", 4); err != nil {
		return false, fmt.Errorf("failed to set Start value: %w", err)
	}

	s.logger.Info("Windows serial mouse Start registry value changed from 3 to 4")
	return true, nil
}

// skipSerialPortEnumeration writes a SkipEnumerations DWORD of 0xffffffff to
// every built-in serial port instance under
// SYSTEM\CurrentControlSet\Enum\ACPI\PNP0501, preventing Windows from polling
// those COM ports for a serial mouse. It returns the instance paths it updated.
func (s *windowsSerialmouseDisable) skipSerialPortEnumeration() ([]string, error) {
	const enumPath = `System\CurrentControlSet\Enum\ACPI\PNP0501`
	const skipValue = 0xffffffff

	parent, err := registry.OpenKey(registry.LOCAL_MACHINE, enumPath, registry.ENUMERATE_SUB_KEYS)
	if errors.Is(err, registry.ErrNotExist) {
		s.logger.Debugf("registry key %s does not exist, skipping serial port enumeration", enumPath)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", enumPath, err)
	}
	defer parent.Close()

	instances, err := parent.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate serial port instances under %s: %w", enumPath, err)
	}

	enumerationDisabledPorts := make([]string, 0, len(instances))
	for _, instance := range instances {
		paramsPath := fmt.Sprintf(`%s\%s\Device Parameters`, enumPath, instance)

		// CreateKey opens the key, creating it (and the Device Parameters subkey)
		// if it does not already exist.
		k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, paramsPath, registry.SET_VALUE)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", paramsPath, err)
		}

		if err := k.SetDWordValue("SkipEnumerations", skipValue); err != nil {
			k.Close()
			return nil, fmt.Errorf("failed to set SkipEnumerations on %s: %w", paramsPath, err)
		}
		k.Close()

		s.logger.Infof("Set SkipEnumerations=0xffffffff on %s", paramsPath)
		enumerationDisabledPorts = append(enumerationDisabledPorts,
			fmt.Sprintf(`%s\SkipEnumerations : 0xffffffff`, paramsPath))
	}

	return enumerationDisabledPorts, nil
}

func (s *windowsSerialmouseDisable) Close(context.Context) error {
	s.cancelFunc()
	return nil
}
