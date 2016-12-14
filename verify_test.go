package httpsignatures

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

// Test Parser

func TestVerifyMissingSignatureShouldFail(t *testing.T) {
	var v VerificationParameters
	err := v.FromString(`keyId="Test",algorithm="hmac-sha256"`)
	assert.Equal(t, "Missing signature", err.Error())
	sigParam := SignatureParameters{KeyID: "Test", Algorithm: AlgorithmHmacSha256, Headers: HeaderList{"date": ""}}
	assert.Equal(t, VerificationParameters{SigParams: &sigParam}, v)
}

func TestVerifyMissingAlgorithmShouldFail(t *testing.T) {
	var v VerificationParameters
	err := v.FromString(`keyId="Test",signature="fffff"`)
	assert.Equal(t, "Missing algorithm", err.Error())
	sigParam := SignatureParameters{KeyID: "Test", Headers: HeaderList{"date": ""}}
	assert.Equal(t, VerificationParameters{SigParams: &sigParam, Signature: "fffff"}, v)
}

func TestVerifyMissingKeyIdShouldFail(t *testing.T) {
	var v VerificationParameters
	err := v.FromString(`algorithm="hmac-sha256",signature="fffff"`)
	assert.Equal(t, "Missing keyId", err.Error())
	sigParam := SignatureParameters{Algorithm: AlgorithmHmacSha256, Headers: HeaderList{"date": ""}}
	assert.Equal(t, VerificationParameters{SigParams: &sigParam, Signature: "fffff"}, v)
}

func TestVerifyDualHeaderShouldPickLastOne(t *testing.T) {
	var v VerificationParameters
	err := v.FromString(`keyId="Test",algorithm="hmac-sha256",signature="fffff",signature="abcde"`)
	assert.Nil(t, err)
	sigParam := SignatureParameters{KeyID: "Test", Algorithm: AlgorithmHmacSha256, Headers: HeaderList{"date": ""}}
	assert.Equal(t, VerificationParameters{SigParams: &sigParam, Signature: "abcde"}, v)
}

func TestVerifyMissingDateHeader(t *testing.T) {
	var v VerificationParameters
	err := v.FromString(`keyId="Test",algorithm="hmac-sha256",signature="fffff",headers="(request-target) host"`)
	assert.Nil(t, err)
	sigParam := SignatureParameters{KeyID: "Test", Algorithm: AlgorithmHmacSha256, Headers: HeaderList{"(request-target)": "", "host": ""}}
	assert.Equal(t, VerificationParameters{SigParams: &sigParam, Signature: "fffff"}, v)
}

func TestVerifyInvalidKeyShouldBeIgnored(t *testing.T) {
	var v VerificationParameters
	err := v.FromString(`keyId="Test",algorithm="hmac-sha256",garbage="bob",signature="fffff"`)
	assert.Nil(t, err)
	sigParam := SignatureParameters{KeyID: "Test", Algorithm: AlgorithmHmacSha256, Headers: HeaderList{"date": ""}}
	assert.Equal(t, VerificationParameters{SigParams: &sigParam, Signature: "fffff"}, v)
}

func TestVerifySignatureFromAuthorizationHeader(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"Date":              []string{testDate},
			HeaderAuthorization: []string{authScheme + testSignature},
		},
	}

	var v VerificationParameters
	err := v.FromRequest(r)
	assert.Nil(t, err)
	assert.Equal(t, "Test", v.SigParams.KeyID)
	assert.Equal(t, AlgorithmHmacSha256, v.SigParams.Algorithm)
	assert.Equal(t, testHash, v.Signature)

	valid, err := v.Verify(testKey, r)
	assert.Nil(t, err)
	assert.Equal(t, true, valid)
}

func TestVerifySignatureFromSignatureHeader(t *testing.T) {
	r := http.Request{
		Header: http.Header{
			"Date":          []string{testDate},
			HeaderSignature: []string{testSignature},
		},
	}

	var v VerificationParameters
	err := v.FromRequest(&r)
	assert.Nil(t, err)

	assert.Equal(t, "Test", v.SigParams.KeyID)
	assert.Equal(t, AlgorithmHmacSha256, v.SigParams.Algorithm)
	assert.Equal(t, testHash, v.Signature)

	valid, err := v.Verify(testKey, &r)
	assert.Nil(t, err)
	assert.Equal(t, true, valid)
}
