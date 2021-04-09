package features

import (
	"os"
	"strings"
)

const INTERPOLATION = "interpolation"

func SetFeature(features map[string]struct{}, f string, val bool) {
	if val {
		features[f] = struct{}{}
	} else {
		delete(features, f)
	}
}

func Features() map[string]struct{} {
	features := map[string]struct{}{}
	// setup defaults
	SetFeature(features, INTERPOLATION, true)
	setting := os.Getenv("SPIFF_FEATURES")
	for _, f := range strings.Split(setting, ",") {
		f = strings.ToLower(strings.TrimSpace(f))
		no := strings.HasPrefix(f, "no")
		if no {
			f = f[2:]
		}
		switch f {
		case INTERPOLATION:
			SetFeature(features, INTERPOLATION, !no)
		}
	}
	return features
}

func EncryptionKey() string {
	return os.Getenv("SPIFF_ENCRYPTION_KEY")
}
