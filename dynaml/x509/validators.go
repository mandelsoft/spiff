package x509

import (
	"fmt"
	"github.com/mandelsoft/spiff/dynaml"
)

func init() {
	dynaml.RegisterValidator("publickey", ValPublicKey)
	dynaml.RegisterValidator("privatekey", ValPrivateKey)
	dynaml.RegisterValidator("certificate", ValCertificate)
	dynaml.RegisterValidator("ca", ValCA)
}
func ValPrivateKey(value interface{}, binding dynaml.Binding, args ...interface{}) (bool, string, string, error, bool) {
	s, err := dynaml.StringValue("privatekey", value)
	if err != nil {
		return dynaml.ValidatorErrorf("%s", err)
	}
	_, err = ParsePrivateKey(s)
	if err != nil {
		return false, "is private key", fmt.Sprintf("is no private key: %s", err), nil, true
	}
	return false, "is private key", "is no private key", nil, true
}

func ValCertificate(value interface{}, binding dynaml.Binding, args ...interface{}) (bool, string, string, error, bool) {
	s, err := dynaml.StringValue("certificate", value)
	if err != nil {
		return dynaml.ValidatorErrorf("%s", err)
	}
	_, err = ParseCertificate(s)
	if err != nil {
		return false, "is certificate", fmt.Sprintf("is no certificate: %s", err), nil, true
	}
	return false, "is certificate", "is no certificate", nil, true
}

func ValCA(value interface{}, binding dynaml.Binding, args ...interface{}) (bool, string, string, error, bool) {
	s, err := dynaml.StringValue("ca", value)
	if err != nil {
		return dynaml.ValidatorErrorf("%s", err)
	}
	c, err := ParseCertificate(s)
	if err != nil {
		return false, "is ca", fmt.Sprintf("is no certificate: %s", err), nil, true
	}
	if !c.IsCA {
		return false, "is ca", fmt.Sprintf("is no ca certificate: %s", err), nil, true
	}
	return false, "is ca", "is no ca", nil, true
}

func ValPublicKey(value interface{}, binding dynaml.Binding, args ...interface{}) (bool, string, string, error, bool) {
	s, err := dynaml.StringValue("publickey", value)
	if err != nil {
		return dynaml.ValidatorErrorf("%s", err)
	}
	_, err = ParsePublicKey(s)
	if err != nil {
		return false, "is public key", fmt.Sprintf("is no public key: %s", err), nil, true
	}
	return false, "is public key", "is no public key", nil, true
}
