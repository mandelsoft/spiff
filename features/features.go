package features

import (
	"os"
	"strings"
)

const INTERPOLATION = "interpolation"
const CONTROL = "control"

type FeatureFlags map[string]struct{}

func (this FeatureFlags) Enabled(name string) bool {
	if this == nil {
		return false
	}
	_, ok := this[name]
	return ok
}
func (this FeatureFlags) Set(name string, active bool) {
	if active {
		this[name] = struct{}{}
	} else {
		delete(this, name)
	}
}

func (this FeatureFlags) InterpolationEnabled() bool {
	return this.Enabled(INTERPOLATION)
}
func (this FeatureFlags) SetInterpolation(active bool) {
	this.Set(INTERPOLATION, active)
}

func (this FeatureFlags) ControlEnabled() bool {
	return this.Enabled(CONTROL)
}
func (this FeatureFlags) SetControl(active bool) {
	this.Set(CONTROL, active)
}

func Features() FeatureFlags {
	features := FeatureFlags{}
	// setup defaults
	features.Set(INTERPOLATION, true)
	setting := os.Getenv("SPIFF_FEATURES")
	for _, f := range strings.Split(setting, ",") {
		f = strings.ToLower(strings.TrimSpace(f))
		no := strings.HasPrefix(f, "no")
		if no {
			f = f[2:]
		}
		switch f {
		case INTERPOLATION:
			features.Set(INTERPOLATION, !no)
		case CONTROL:
			features.Set(CONTROL, !no)
		}
	}
	return features
}

func EncryptionKey() string {
	return os.Getenv("SPIFF_ENCRYPTION_KEY")
}
