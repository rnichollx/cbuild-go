package system

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

type Platform int

const (
	PlatformUnknown Platform = iota
	PlatformWindows
	PlatformMac
	PlatformLinux
	PlatformFreeBSD
)

func (p Platform) String() string {
	switch p {
	case PlatformWindows:
		return "Windows"
	case PlatformMac:
		return "Mac"
	case PlatformLinux:
		return "Linux"
	case PlatformFreeBSD:
		return "FreeBSD"
	}
	return "Unknown"
}

func (p Platform) StringLower() string {
	return strings.ToLower(p.String())
}

func (p Platform) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

func (p *Platform) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "windows":
		*p = PlatformWindows
	case "mac", "macos", "darwin":
		*p = PlatformMac
	case "linux":
		*p = PlatformLinux
	case "freebsd":
		*p = PlatformFreeBSD
	case "unknown":
		*p = PlatformUnknown
	default:
		return errors.New("Platform.UnmarshalYAML: unrecognized platform: " + s)
	}
	return nil
}
