// Code generated by counterfeiter. DO NOT EDIT.
package mock

import (
	"sync"

	"github.com/hyperledger-labs/fabric-token-sdk/token/core/zkatdlog/crypto/audit"
)

type SigningIdentity struct {
	SignStub        func(raw []byte) ([]byte, error)
	signMutex       sync.RWMutex
	signArgsForCall []struct {
		raw []byte
	}
	signReturns struct {
		result1 []byte
		result2 error
	}
	signReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	VerifyStub        func(message, sigma []byte) error
	verifyMutex       sync.RWMutex
	verifyArgsForCall []struct {
		message []byte
		sigma   []byte
	}
	verifyReturns struct {
		result1 error
	}
	verifyReturnsOnCall map[int]struct {
		result1 error
	}
	SerializeStub        func() ([]byte, error)
	serializeMutex       sync.RWMutex
	serializeArgsForCall []struct{}
	serializeReturns     struct {
		result1 []byte
		result2 error
	}
	serializeReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *SigningIdentity) Sign(raw []byte) ([]byte, error) {
	var rawCopy []byte
	if raw != nil {
		rawCopy = make([]byte, len(raw))
		copy(rawCopy, raw)
	}
	fake.signMutex.Lock()
	ret, specificReturn := fake.signReturnsOnCall[len(fake.signArgsForCall)]
	fake.signArgsForCall = append(fake.signArgsForCall, struct {
		raw []byte
	}{rawCopy})
	fake.recordInvocation("Sign", []interface{}{rawCopy})
	fake.signMutex.Unlock()
	if fake.SignStub != nil {
		return fake.SignStub(raw)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.signReturns.result1, fake.signReturns.result2
}

func (fake *SigningIdentity) SignCallCount() int {
	fake.signMutex.RLock()
	defer fake.signMutex.RUnlock()
	return len(fake.signArgsForCall)
}

func (fake *SigningIdentity) SignArgsForCall(i int) []byte {
	fake.signMutex.RLock()
	defer fake.signMutex.RUnlock()
	return fake.signArgsForCall[i].raw
}

func (fake *SigningIdentity) SignReturns(result1 []byte, result2 error) {
	fake.SignStub = nil
	fake.signReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *SigningIdentity) SignReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.SignStub = nil
	if fake.signReturnsOnCall == nil {
		fake.signReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.signReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *SigningIdentity) Verify(message []byte, sigma []byte) error {
	var messageCopy []byte
	if message != nil {
		messageCopy = make([]byte, len(message))
		copy(messageCopy, message)
	}
	var sigmaCopy []byte
	if sigma != nil {
		sigmaCopy = make([]byte, len(sigma))
		copy(sigmaCopy, sigma)
	}
	fake.verifyMutex.Lock()
	ret, specificReturn := fake.verifyReturnsOnCall[len(fake.verifyArgsForCall)]
	fake.verifyArgsForCall = append(fake.verifyArgsForCall, struct {
		message []byte
		sigma   []byte
	}{messageCopy, sigmaCopy})
	fake.recordInvocation("Verify", []interface{}{messageCopy, sigmaCopy})
	fake.verifyMutex.Unlock()
	if fake.VerifyStub != nil {
		return fake.VerifyStub(message, sigma)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.verifyReturns.result1
}

func (fake *SigningIdentity) VerifyCallCount() int {
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	return len(fake.verifyArgsForCall)
}

func (fake *SigningIdentity) VerifyArgsForCall(i int) ([]byte, []byte) {
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	return fake.verifyArgsForCall[i].message, fake.verifyArgsForCall[i].sigma
}

func (fake *SigningIdentity) VerifyReturns(result1 error) {
	fake.VerifyStub = nil
	fake.verifyReturns = struct {
		result1 error
	}{result1}
}

func (fake *SigningIdentity) VerifyReturnsOnCall(i int, result1 error) {
	fake.VerifyStub = nil
	if fake.verifyReturnsOnCall == nil {
		fake.verifyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.verifyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *SigningIdentity) Serialize() ([]byte, error) {
	fake.serializeMutex.Lock()
	ret, specificReturn := fake.serializeReturnsOnCall[len(fake.serializeArgsForCall)]
	fake.serializeArgsForCall = append(fake.serializeArgsForCall, struct{}{})
	fake.recordInvocation("Serialize", []interface{}{})
	fake.serializeMutex.Unlock()
	if fake.SerializeStub != nil {
		return fake.SerializeStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.serializeReturns.result1, fake.serializeReturns.result2
}

func (fake *SigningIdentity) SerializeCallCount() int {
	fake.serializeMutex.RLock()
	defer fake.serializeMutex.RUnlock()
	return len(fake.serializeArgsForCall)
}

func (fake *SigningIdentity) SerializeReturns(result1 []byte, result2 error) {
	fake.SerializeStub = nil
	fake.serializeReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *SigningIdentity) SerializeReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.SerializeStub = nil
	if fake.serializeReturnsOnCall == nil {
		fake.serializeReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.serializeReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *SigningIdentity) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.signMutex.RLock()
	defer fake.signMutex.RUnlock()
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	fake.serializeMutex.RLock()
	defer fake.serializeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *SigningIdentity) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ audit.SigningIdentity = new(SigningIdentity)