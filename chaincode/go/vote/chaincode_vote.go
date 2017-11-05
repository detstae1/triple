
package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strings"
	"encoding/pem"
	"crypto/x509"
	"encoding/json"
	"bytes"
	"encoding/gob"
)

var logger = shim.NewLogger("VoteChaincode")


type VoteChaincode struct {
}

type VoteKey struct {
	Question string
	Org string
	User string
}

type Vote struct {
	Key VoteKey
	Answer string
}

func (t *VoteChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	allowedOrgs := stub.GetStringArgs()

	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(allowedOrgs)
	bs := buf.Bytes()

	stub.PutState("allowedOrgs", bs)

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

	// check if allowed
	bs, err := stub.GetState("allowedOrgs")
	if err != nil {
		return shim.Error("cannot get allowed orgs")
	}
	allowedOrgs := []string{}
	gob.NewDecoder(bytes.NewReader(bs)).Decode(&allowedOrgs)

	if !contains(allowedOrgs, org) {
		return shim.Error("not allowed to vote!")
	}

	key, err := stub.CreateCompositeKey("Vote", []string{question, org, user})

	if err != nil {
		return shim.Error("cannot CreateCompositeKey")
	}

	stub.PutState(key, []byte(answer))

	return shim.Success(nil)
}

func (t *VoteChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var keys []string

	if len(args) > 2 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	} else if len(args) == 1 {
		question := args[0]
		keys = []string{question}
	} else if len(args) == 2 {
		question := args[0]
		org := args[1]
		keys = []string{question, org}
	}

	it, err := stub.GetStateByPartialCompositeKey("Vote", keys)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	arr := []Vote{}
	for it.HasNext() {
		next, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		/*var voteValue string
		err = json.Unmarshal(next.Value, &voteValue)
		if err != nil {
			return shim.Error(err.Error())
		}*/

		voteValue := string(next.Value)

		_, keys, err := stub.SplitCompositeKey(next.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		voteKey := VoteKey{Question: keys[0], Org: keys[1]}

		vote := Vote{Key: voteKey, Answer: voteValue}

		arr = append(arr, vote)
	}

	ret, err := json.Marshal(arr)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ret)
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func main() {
	err := shim.Start(new(VoteChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
