# Module windows-serialmouse

Microsoft Windows sometimes it detects other serial devices as a serial mouse. If the device is connected to a real TIA-232 port or a Universal Serial Bus (USB) to TIA-232 adapter. Some USB connected devices, e.g. GPS receivers, have USB to TIA-232 convertors built right into the device.  This Viam module disables the serial mouse if Windows has re-enabled it.

<https://paulhutch.blog/2019/06/24/disable-serial-mouse-detection/>

## Models

This module provides the following model:

- Model viam-soleng:windows-serialmouse:disable

### windows-serialmouse

Call the DoCommand of this model to apply two registry changes that stop Windows from treating the serial line as a serial mouse:

1. Disable the `sermouse` service. The module queries the registry and, if `Start` is 3, changes it to 4.

   ```text
   Location: HKEY_LOCAL_MACHINE\System\CurrentControlSet\Services\sermouse
   Key: Start
   Value: 3 -> 4
   ```

2. Skip serial port enumeration. For every built-in serial port instance under `PNP0501`, the module writes a `SkipEnumerations` DWORD of `0xffffffff` so Windows stops polling those COM ports for a mouse. The `Device Parameters` subkey and value are created if they do not already exist.

   ```text
   Location: HKEY_LOCAL_MACHINE\System\CurrentControlSet\Enum\ACPI\PNP0501\<instance>\Device Parameters
   Key: SkipEnumerations (DWORD)
   Value: 0xffffffff
   ```

   `PNP0501` is the ACPI ID for the built-in 16550 serial port. A device on a USB-to-serial adapter or a different ACPI ID will not be covered by this step.

## Configuration

There are no configuration attributes required for this module.

### Attributes

There are no configuration attributes required for this module.

### Example Configuration

```json
{
}
```

## DoCommand

This module implements the DoCommand. Periodically send it an empty `{}` via the Viam Job Scheduler.

### Example DoCommand

```json
{
}
```

### Response

```json
{
  "start": 4,
  "changed": true,
  "message": "Start value changed from 3 to 4",
  "skipped_ports": [
    "System\\CurrentControlSet\\Enum\\ACPI\\PNP0501\\1\\Device Parameters"
  ]
}
```

- `changed` — whether the `sermouse` `Start` value was updated this call (`false` if it was already 4).
- `skipped_ports` — the `Device Parameters` paths where `SkipEnumerations` was written. Empty if no `PNP0501` serial port instances were found.
