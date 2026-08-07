package dynaml

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/md4"
)

const JCS_INDICATOR = "*"

func func_md5(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("md5 takes exactly one argument")
	}

	str, ok := arguments[0].(string)
	if !ok {
		s, err := CanonicalizedJson(arguments[0])
		if err != nil {
			return info.Error("%s", err)
		}
		str = string(s)
	}

	result := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", result), info, true
}

func func_hash(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	var err error

	info := DefaultInfo()

	if len(arguments) < 1 || len(arguments) > 2 {
		return info.Error("hash takes one or two arguments")
	}

	mode := ""

	if len(arguments) == 2 {
		str, ok := arguments[1].(string)
		if !ok {
			return info.Error("second argument for hash must be a string")
		}
		mode = str
	}

	jcs := false
	bin := false
	if strings.HasPrefix(mode, JCS_INDICATOR) {
		mode = mode[len(JCS_INDICATOR):]
		jcs = true
	} else if strings.HasPrefix(mode, BINARY_INDICATOR) {
		mode = mode[len(BINARY_INDICATOR):]
		bin = true
	}

	if mode == "" {
		mode = "sha256"
	}

	var data []byte

	str, ok := arguments[0].(string)
	if !ok && bin {
		return info.Error("first argument for a binary hash must be a string")
	}
	if jcs || !ok {
		data, err = CanonicalizedJson(arguments[0])
		if err != nil {
			return info.Error("%s", err)
		}
	} else {
		data = []byte(str)
	}

	if bin {
		var err error
		data, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return info.Error("cannot decode base64 string: %s", err)
		}
	}

	var result []byte

	switch mode {
	case "md4":
		result = md4.New().Sum(data)
	case "md5":
		r := md5.Sum(data)
		result = r[:]
	case "sha1":
		r := sha1.Sum(data)
		result = r[:]
	case "sha224":
		r := sha256.Sum224(data)
		result = r[:]
	case "sha256":
		r := sha256.Sum256(data)
		result = r[:]
	case "sha384":
		r := sha512.Sum384(data)
		result = r[:]
	case "sha512":
		r := sha512.Sum512(data)
		result = r[:]
	case "sha512/224":
		r := sha512.Sum512_224(data)
		result = r[:]
	case "sha512/256":
		r := sha512.Sum512_256(data)
		result = r[:]
	default:
		return info.Error("invalid hash type '%s'", mode)
	}
	return fmt.Sprintf("%x", result), info, true
}
