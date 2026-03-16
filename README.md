# Module windows-serialmouse

Microsoft Windows sometimes it detects other serial devices as a serial mouse. If the device is connected to a real TIA-232 port or a Universal Serial Bus (USB) to TIA-232 adapter. Some USB connected devices, e.g. GPS receivers, have USB to TIA-232 convertors built right into the device.  This Viam module disables the serial mouse if Windows has re-enabled it.

<https://paulhutch.blog/2019/06/24/disable-serial-mouse-detection/>

## Models

This module provides the following model:

- Model viam-soleng:windows-serialmouse:disable

### windows-serialmouse

Call the DoCommand of this model to query the Windows registry and if this HKEY is 3, change it to 4.

```text
Location: HKEY_LOCAL_MACHINE\System\CurrentControlSet\Services\sermouse
Key: Start 
Value: 3
```

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
