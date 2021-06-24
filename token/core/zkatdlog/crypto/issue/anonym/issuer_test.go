/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package anonym_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hyperledger-labs/fabric-token-sdk/token/core/math/gurvy/bn256"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/zkatdlog/crypto"
	issue2 "github.com/hyperledger-labs/fabric-token-sdk/token/core/zkatdlog/crypto/issue"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/zkatdlog/crypto/issue/anonym"
)

var _ = Describe("Issuer", func() {
	var (
		pp     *crypto.PublicParams
		issue  *issue2.IssueAction
		issuer *anonym.Issuer
		signer *anonym.Signer
		values []uint64
		bf     []*bn256.Zr
		owners [][]byte
	)
	BeforeEach(func() {
		var err error
		owners = make([][]byte, 3)
		owners[0] = []byte("alice")
		owners[1] = []byte("bob")
		owners[2] = []byte("charlie")

		values = []uint64{50, 30, 20}

		bf = make([]*bn256.Zr, 3)
		rand, err := bn256.GetRand()
		Expect(err).NotTo(HaveOccurred())
		for i := 0; i < 3; i++ {
			bf[i] = bn256.RandModOrder(rand)
		}

		pp, err = crypto.Setup(100, 2, nil)
		Expect(err).NotTo(HaveOccurred())

		sk, pk, err := anonym.GenerateKeyPair("ABC", pp)
		Expect(err).NotTo(HaveOccurred())

		issuers := GetIssuers(2, 1, pk, pp.ZKATPedParams)
		err = pp.SetIssuingPolicy(issuers)
		Expect(err).NotTo(HaveOccurred())

		witness := anonym.NewWitness(sk, nil, nil, nil, nil, 1)
		signer = anonym.NewSigner(witness, nil, nil, 1, pp.ZKATPedParams)
		issuer = &anonym.Issuer{}
		issuer.New("ABC", signer, pp)

	})

	Describe("Issue", func() {
		When("issue is computed correctly", func() {
			It("succeeds", func() {
				var err error
				issue, _, err = issuer.GenerateZKIssue(values, owners)
				Expect(err).NotTo(HaveOccurred())
				Expect(issue).NotTo(BeNil())
				raw, err := issue.Serialize()
				Expect(err).NotTo(HaveOccurred())
				sig, err := issuer.SignTokenActions(raw, "0")
				Expect(err).NotTo(HaveOccurred())
				signed := append(raw, []byte("0")...)
				err = issuer.Signer.Verify(signed, sig)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})