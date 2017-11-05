
package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strings"
	"encoding/pem"
	"crypto/x509"
)

var logger = shim.NewLogger("VoteChaincode")

type VoteChaincode struct {
}

type VoteKey struct {
	question string
	org string
	user string
}

type Vote struct {
	key VoteKey
	answer uint8
}

func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *VoteChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "cast" {
		return t.cast(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:400,Message:"invalid function name"}
}

func (t *VoteChaincode) cast(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	question := args[0]
	answer := args[1]

	creatorBytes, err := stub.GetCreator()

	if err != nil {
		return shim.Error("cannot GetCreator")
	}

	user, org := getCreator(creatorBytes)

	key, err := stub.CreateCompositeKey("Vote", []string{question, org, user})

	if err != nil {
		return shim.Error("cannot CreateCompositeKey")
	}

	stub.PutState(key, []byte(answer))

	return shim.Success(nil)
}

func (t *VoteChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	return shim.Success(nil)
}

var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
