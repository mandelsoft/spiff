package x509

import (
	"bufio"
	"bytes"
	"encoding/pem"
	. "github.com/mandelsoft/spiff/dynaml"
)

const F_PublicKey = "x509publickey"

func init() {
	RegisterFunction(F_PublicKey, func_x509publickey)
}

// one argument
//  - private key pem

func func_x509publickey(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	var err error
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("invalid argument count for %s(<privatekey>)", F_PublicKey)
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("argument for %s must be a private key in pem format", F_PublicKey)
	}

	key, err := ParsePrivateKey(str)
	if err != nil {
		return info.Error("argument for %s must be a private key in pem format: %s", F_PublicKey, err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	if err := pem.Encode(writer, pemBlockForPublicKey(publicKey(key))); err != nil {
		return info.Error("failed to write public key pem block: %s", err)
	}
	writer.Flush()
	return b.String(), info, true
}
