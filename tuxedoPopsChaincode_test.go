package main

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	txcache "github.com/skuchain/TuxedoPops/TXCache"
	"github.com/skuchain/TuxedoPops/TuxedoPopsTX"

	"encoding/hex"
	"encoding/json"

	"github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// Notes fromessage Testing popcode
// Public Key: 02ca4a8c7dc5090f924cde2264af240d76f6d58a5d2d15c8c5f59d95c70bd9e4dc
// Private Key: 94d7fe7308a452fdf019a0424d9c48ba9b66bdbca565c6fa3b1bf9c646ebac20
// Hyperledger address hex 74ded2036e988fc56e3cff77a40c58239591e921
// Hyperledger address Base58: 8sDMfw2Ti7YumfTkbf7RHMgSSSxuAmMFd2GS9wnjkUoX

// Notes fromessage Testing popcode2
// Public Key: 02cb6d65b04c4b84502015f918fe549e95cad4f3b899359a170d4d7d438363c0ce
// Private Key: 60977f22a920c9aa18d58d12cb5e90594152d7aa724bcce21484dfd0f4490b58
// Hyperledger address hex 10734390011641497f489cb475743b8e50d429bb
// Hyperledger address Base58: EHxhLN3Ft4p9jPkR31MJMEMee9G

//Owner1 key
// Public Key: 0278b76afbefb1e1185bc63ed1a17dd88634e0587491f03e9a8d2d25d9ab289ee7
// Private Key: 7142c92e6eba38de08980eeb55b8c98bb19f8d417795adb56b6c4d25da6b26c5

// Owner2 key
// Public Key: 02e138b25db2e74c54f8ca1a5cf79e2d1ed6af5bd1904646e7dc08b6d7b0d12bfd
// Private Key: b18b7d3082b3ff9438a7bf9f5f019f8a52fb64647ea879548b3ca7b551eefd65

func checkInit(t *testing.T, stub *shim.MockStub, args []string) {
	_, err := stub.MockInit("1", "", args)
	if err != nil {
		HandleError(t, fmt.Errorf("INIT", args, "failed", err))
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args []string) {
	_, err := stub.MockInvoke("1", "invoke", args)
	if err != nil {
		HandleError(t, fmt.Errorf("invoke", args, "failed", err))
		t.FailNow()
	}
}

func checkQuery(t *testing.T, stub *shim.MockStub, name string, value string) {
	bytes, err := stub.MockQuery("balance", []string{name})

	if err != nil {
		HandleError(t, fmt.Errorf("Query for address (", name, ") failed", err))
		t.FailNow()
	}
	if bytes == nil {
		HandleError(t, fmt.Errorf("Query for address (", name, ") failed to get value"))
		t.FailNow()
	}
	if string(bytes) != value {
		HandleError(t, fmt.Errorf("Query value for address (%s) wanted:\n(%s)\n\nGot:\n(%s)\n", name, value, string(bytes)))
		t.FailNow()
	}
}

//mint performs a create transaction with all parameters (other than counterseed) hard coded rather than taken as arguments.
func mint(t *testing.T, stub *shim.MockStub, counterSeed string) {
	createArgs := TuxedoPopsTX.CreateTX{}
	createArgs.Address = "74ded2036e988fc56e3cff77a40c58239591e921"
	createArgs.Amount = 10
	pubKeyBytes, err := hex.DecodeString("03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26")
	if err != nil {
		HandleError(t, err)
	}
	createArgs.CreatorPubKey = pubKeyBytes
	hexCreatorSig := generateCreateSig(counterSeed, 10, "Test Asset", "Test Data",
		"74ded2036e988fc56e3cff77a40c58239591e921", "7ff1ac3d9dfc56315ee610d0a15609d13c399cf9c92ba2e32e7b1d25ea5c9494")

	createArgs.CreatorSig, err = hex.DecodeString(hexCreatorSig)
	if err != nil {
		HandleError(t, err)
	}
	createArgs.Data = "Test Data"
	createArgs.Type = "Test Asset"
	createArgBytes, err := proto.Marshal(&createArgs)
	createArgBytesStr := hex.EncodeToString(createArgBytes)
	_, err = stub.MockInvoke("3", "create", []string{createArgBytesStr})
	if err != nil {
		HandleError(t, err)
	}
}

//altMint takes in a key struct which holds input of private key, public key, address, and counter
//it hard codes remaining parameters
func altMint(t *testing.T, stub *shim.MockStub, keys *keyInfo) {
	createArgs := TuxedoPopsTX.CreateTX{}
	createArgs.Address = keys.address
	createArgs.Amount = 10
	createArgs.Data = "Test Data"
	createArgs.Type = "Test Asset"

	pubKeyBytes, err := hex.DecodeString(keys.pubKeyStr)
	if err != nil {
		HandleError(t, err)
	}
	createArgs.CreatorPubKey = pubKeyBytes
	hexCreatorSig := generateCreateSig(keys.counter, 10, "Test Asset", "Test Data", keys.address, keys.privKeyStr)

	createArgs.CreatorSig, err = hex.DecodeString(hexCreatorSig)
	if err != nil {
		HandleError(t, err)
	}

	createArgBytes, err := proto.Marshal(&createArgs)
	createArgBytesStr := hex.EncodeToString(createArgBytes)
	_, err = stub.MockInvoke("3", "create", []string{createArgBytesStr})
	if err != nil {
		HandleError(t, err)
	}
}

//altMint1 does not hard code any parameters, taking them all as arguments
func altMint1(t *testing.T, stub *shim.MockStub, user *keyInfo, popcode *keyInfo,
	data string, createdType string, amount int) {

	createArgs := TuxedoPopsTX.CreateTX{}
	createArgs.Address = popcode.address
	createArgs.Amount = int32(amount)
	createArgs.Data = data
	createArgs.Type = createdType

	creatorPubKeyBytes, err := hex.DecodeString(user.pubKeyStr)
	if err != nil {
		HandleError(t, err)
	}
	createArgs.CreatorPubKey = creatorPubKeyBytes
	popcode.counter, err = getCounter(stub, popcode)
	if err != nil {
		HandleError(t, fmt.Errorf("Error getting counterseed in altMint: (%s)", err.Error()))
		t.FailNow()
	}
	hexCreatorSig := generateCreateSig(popcode.counter, amount, createdType, data, popcode.address, user.privKeyStr)

	createArgs.CreatorSig, err = hex.DecodeString(hexCreatorSig)
	if err != nil {
		HandleError(t, err)
	}

	createArgBytes, err := proto.Marshal(&createArgs)
	createArgBytesStr := hex.EncodeToString(createArgBytes)
	_, err = stub.MockInvoke("3", "create", []string{createArgBytesStr})
	if err != nil {
		HandleError(t, err)
	}
}

func generateCreateSig(CounterSeedStr string, amount int, assetType string,
	data string, addr string, privateKeyStr string) string {

	privKeyByte, _ := hex.DecodeString(privateKeyStr)

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyByte)

	message := CounterSeedStr + ":" + addr + ":" +
		strconv.FormatInt(int64(amount), 10) + ":" + assetType + ":" + data
	fmt.Println("Signed Message")
	fmt.Println(message)
	messageBytes := sha256.Sum256([]byte(message))
	sig, _ := privKey.Sign(messageBytes[:])
	return hex.EncodeToString(sig.Serialize())
}

func possess(t *testing.T, stub *shim.MockStub, counterSeed string, idx int) {
	transferArgs := TuxedoPopsTX.TransferOwners{}
	transferArgs.Address = "74ded2036e988fc56e3cff77a40c58239591e921"
	transferArgs.Data = "Test possess"
	transferArgs.PopcodePubKey, _ = hex.DecodeString("02ca4a8c7dc5090f924cde2264af240d76f6d58a5d2d15c8c5f59d95c70bd9e4dc")
	ownerBytes, _ := hex.DecodeString("0278b76afbefb1e1185bc63ed1a17dd88634e0587491f03e9a8d2d25d9ab289ee7")
	transferArgs.Owners = [][]byte{ownerBytes}
	transferArgs.Output = int32(idx)
	transferArgs.Threshold = 0
	ownerHex := hex.EncodeToString(ownerBytes)

	hexPossessSig := generatePossessSig(t, counterSeed, idx, transferArgs.Threshold, "Test possess",
		ownerHex, "94d7fe7308a452fdf019a0424d9c48ba9b66bdbca565c6fa3b1bf9c646ebac20")

	var err error
	transferArgs.PopcodeSig, err = hex.DecodeString(hexPossessSig)
	transferArgsBytes, _ := proto.Marshal(&transferArgs)
	transferArgsBytesStr := hex.EncodeToString(transferArgsBytes)

	_, err = stub.MockInvoke("4", "transfer", []string{transferArgsBytesStr})
	if err != nil {
		HandleError(t, err)
		t.FailNow()
	}
}

func altPossess(t *testing.T, stub *shim.MockStub, popcode *keyInfo,
	prevOwners []*keyInfo, newOwners []*keyInfo, idx int, data string, threshold int) {

	transferArgs := TuxedoPopsTX.TransferOwners{}
	transferArgs.Address = popcode.address
	transferArgs.Data = data
	transferArgs.PopcodePubKey, _ = hex.DecodeString(popcode.pubKeyStr)
	transferArgs.Owners = [][]byte{}
	transferArgs.Threshold = int32(threshold)
	for _, owner := range newOwners {
		ownerBytes, err := hex.DecodeString(owner.pubKeyStr)
		if err != nil {
			HandleError(t, fmt.Errorf("error decoding public key string (%s) for user address (%s)\n",
				owner.pubKeyStr, owner.address))
			t.FailNow()
		}
		transferArgs.Owners = append(transferArgs.Owners, ownerBytes)
	}
	var err error
	popcode.counter, err = getCounter(stub, popcode)
	if err != nil {
		HandleError(t, fmt.Errorf("error getting counter in altPossess: (%v)\n", err.Error()))
		t.FailNow()
	}
	newOwnersSlice := []string{}
	for _, owner := range transferArgs.Owners {
		newOwnersSlice = append(newOwnersSlice, hex.EncodeToString(owner))
	}
	newOwnersString := strings.Join(newOwnersSlice, ",")
	transferArgs.PrevOwnerSigs = make([][]byte, len(prevOwners))
	for i, owner := range prevOwners {
		if owner != nil {
			hexPrevOwnerSig := generatePossessSig(t, popcode.counter, idx,
				transferArgs.Threshold, data, newOwnersString, owner.privKeyStr)

			transferArgs.PrevOwnerSigs[i], err = hex.DecodeString(hexPrevOwnerSig)
			if err != nil {
				HandleError(t, fmt.Errorf("error decoding hexPrevOwnerSig:\ni=(%d)\nowner = (%v)\nerr: (%v)\n",
					i, owner, err.Error()))
				t.FailNow()
			}

		}
	}
	transferArgs.Output = int32(idx)
	hexPossessSig := generatePossessSig(t, popcode.counter, idx, transferArgs.Threshold, data, newOwnersString, popcode.privKeyStr)
	transferArgs.PopcodeSig, err = hex.DecodeString(hexPossessSig)
	transferArgsBytes, _ := proto.Marshal(&transferArgs)
	transferArgsBytesStr := hex.EncodeToString(transferArgsBytes)

	_, err = stub.MockInvoke("4", "transfer", []string{transferArgsBytesStr})
	if err != nil {
		HandleError(t, fmt.Errorf("POSSESS ERROR: (%v)", err.Error()))
		t.FailNow()
	}
}

func generatePossessSig(t *testing.T, CounterSeedStr string, outputIdx int,
	threshold int32, data string, newOwnersHex string, privateKeyStr string) string {

	privKeyBytes, _ := hex.DecodeString(privateKeyStr)

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)

	message := CounterSeedStr + ":" + strconv.FormatInt(int64(outputIdx), 10)
	if threshold > 0 {
		message += ":" + strconv.FormatInt(int64(threshold), 10)
	}
	message += ":" + data

	newOwnersStrings := strings.Split(newOwnersHex, ",")
	for _, owner := range newOwnersStrings {
		message += ":"
		message += owner
	}
	fmt.Printf("\n\n\nSigned POSSESS message %s \n\n\n", message)
	mDigest := sha256.Sum256([]byte(message))
	sig, _ := privKey.Sign(mDigest[:])
	return hex.EncodeToString(sig.Serialize())
}

// func generatePossessSig(CounterSeedStr string, outputIdx int, data string, newOwners [][]byte, privateKeyStr string) []byte {
// 	privKeyByte, _ := hex.DecodeString(privateKeyStr)
// 	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyByte)
// 	message := CounterSeedStr + ":" + strconv.FormatInt(int64(outputIdx), 10) + ":" + data
// 	for _, newO := range newOwners {
// 		message += ":"
// 		message += hex.EncodeToString(newO)
// 	}
// 	// fmt.Printf("Signed message %s \n", message)
// 	mDigest := sha256.Sum256([]byte(message))
// 	sig, _ := privKey.Sign(mDigest[:])
// 	return sig.Serialize()
// }

func unitize(t *testing.T, stub *shim.MockStub, counterSeed string) {
	unitizeArgs := TuxedoPopsTX.Unitize{}
	unitizeArgs.Data = "Test Unitize"
	unitizeArgs.DestAddress = "10734390011641497f489cb475743b8e50d429bb"
	unitizeArgs.DestAmounts = []int32{10}
	unitizeArgs.SourceAddress = "74ded2036e988fc56e3cff77a40c58239591e921"
	unitizeArgs.SourceOutput = 0
	unitizeArgs.PopcodePubKey, _ = hex.DecodeString("02ca4a8c7dc5090f924cde2264af240d76f6d58a5d2d15c8c5f59d95c70bd9e4dc")
	ownerSig := generateUnitizeSig(counterSeed, unitizeArgs.DestAddress, 0, []int{10}, unitizeArgs.Data,
		"7142c92e6eba38de08980eeb55b8c98bb19f8d417795adb56b6c4d25da6b26c5")

	unitizeArgs.OwnerSigs = [][]byte{ownerSig}
	unitizeArgs.PopcodeSig = generateUnitizeSig(counterSeed, unitizeArgs.DestAddress, 0, []int{10},
		unitizeArgs.Data, "94d7fe7308a452fdf019a0424d9c48ba9b66bdbca565c6fa3b1bf9c646ebac20")

	unitizeArgsBytes, _ := proto.Marshal(&unitizeArgs)
	unitizeArgsBytesStr := hex.EncodeToString(unitizeArgsBytes)

	_, err := stub.MockInvoke("4", "unitize", []string{unitizeArgsBytesStr})
	if err != nil {
		HandleError(t, err)
		t.FailNow()
	}
}

func altUnitize(t *testing.T, stub *shim.MockStub, sourcePopcode *keyInfo,
	destPopcode *keyInfo, owners []*keyInfo, data string, amounts []int32, output int32) {

	unitizeArgs := TuxedoPopsTX.Unitize{}
	unitizeArgs.Data = data
	unitizeArgs.DestAddress = destPopcode.address
	unitizeArgs.DestAmounts = amounts

	unitizeArgs.SourceAddress = sourcePopcode.address
	unitizeArgs.SourceOutput = output
	unitizeArgs.PopcodePubKey, _ = hex.DecodeString(sourcePopcode.pubKeyStr)
	var err error
	sourcePopcode.counter, err = getCounter(stub, sourcePopcode)
	if err != nil {
		HandleError(t, fmt.Errorf("Error getting counter: (%v)\n", err.Error()))
		t.FailNow()
	}
	intAmounts := make([]int, len(amounts))
	for i, amount := range amounts {
		intAmounts[i] = int(amount)
	}

	unitizeArgs.OwnerSigs = [][]byte{}
	for _, owner := range owners {
		ownerSig := generateUnitizeSig(sourcePopcode.counter, unitizeArgs.DestAddress,
			int(output), intAmounts, unitizeArgs.Data, owner.privKeyStr)

		unitizeArgs.OwnerSigs = append(unitizeArgs.OwnerSigs, ownerSig)
	}

	unitizeArgs.PopcodeSig = generateUnitizeSig(sourcePopcode.counter, unitizeArgs.DestAddress,
		int(output), intAmounts, unitizeArgs.Data, sourcePopcode.privKeyStr)

	unitizeArgsBytes, _ := proto.Marshal(&unitizeArgs)
	unitizeArgsBytesStr := hex.EncodeToString(unitizeArgsBytes)

	_, err = stub.MockInvoke("4", "unitize", []string{unitizeArgsBytesStr})
	if err != nil {
		HandleError(t, err)
		t.FailNow()
	}
}

func generateUnitizeSig(CounterSeedStr string, destAddr string, outputIdx int,
	amounts []int, data string, privateKeyStr string) []byte {

	privKeyByte, _ := hex.DecodeString(privateKeyStr)
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyByte)
	message := CounterSeedStr + ":" + destAddr + ":" + data + ":" + strconv.FormatInt(int64(outputIdx), 10)
	for _, amount := range amounts {
		message += ":" + strconv.FormatInt(int64(amount), 10)
	}
	fmt.Printf("\n\nunitize message: (%s)\n\n", message)

	mDigest := sha256.Sum256([]byte(message))
	sig, _ := privKey.Sign(mDigest[:])
	return sig.Serialize()
}

func generateRecipeSig(recipeName string, createdType string,
	ingredients []*TuxedoPopsTX.Ingredient, privateKeyStr string) string {

	privKeyByte, _ := hex.DecodeString(privateKeyStr)

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyByte)

	message := recipeName + ":" + createdType
	for _, ingredient := range ingredients {
		message += ":" + strconv.FormatInt(int64(ingredient.Numerator), 10) + ":" +
			strconv.FormatInt(int64(ingredient.Denominator), 10) + ":" + ingredient.Type
	}
	fmt.Printf("Signed Message: (%s)\n\n\n", message)
	messageBytes := sha256.Sum256([]byte(message))
	sig, _ := privKey.Sign(messageBytes[:])
	return hex.EncodeToString(sig.Serialize())
}

func altRecipe(t *testing.T, stub *shim.MockStub, recipeName string, createdType string,
	creator *keyInfo, ingredients []*TuxedoPopsTX.Ingredient) {

	recipeArgs := TuxedoPopsTX.Recipe{}
	recipeArgs.RecipeName = recipeName
	recipeArgs.CreatedType = createdType
	var err error
	recipeArgs.CreatorPubKey, err = hex.DecodeString(creator.pubKeyStr)
	if err != nil {
		HandleError(t, err)
		t.FailNow()
	}
	recipeArgs.Ingredients = ingredients

	sigHex := generateRecipeSig(recipeArgs.RecipeName, recipeArgs.CreatedType,
		recipeArgs.Ingredients, creator.privKeyStr)

	recipeArgs.CreatorSig, err = hex.DecodeString(sigHex)
	if err != nil {
		HandleError(t, fmt.Errorf("error decoding creator signature in register recipe. ERR: (%v)", err.Error()))
		t.FailNow()
	}
	recipeArgsBytes, err := proto.Marshal(&recipeArgs)
	if err != nil {
		HandleError(t, fmt.Errorf("error marshalling recipeArgs in registerRecipe. ERR: (%s)\n", err.Error()))
		t.FailNow()
	}
	recipeArgsBytesStr := hex.EncodeToString(recipeArgsBytes)
	_, err = stub.MockInvoke("4", "recipe", []string{recipeArgsBytesStr})
	if err != nil {
		HandleError(t, fmt.Errorf("error invoking recipe: (%v)", err.Error()))
		t.FailNow()
	}
}

func registerRecipe(t *testing.T, stub *shim.MockStub) {
	recipeArgs := TuxedoPopsTX.Recipe{}
	recipeArgs.RecipeName = "test recipe"
	recipeArgs.CreatedType = "water vapor"
	recipeArgs.CreatorPubKey, _ = hex.DecodeString("02ca4a8c7dc5090f924cde2264af240d76f6d58a5d2d15c8c5f59d95c70bd9e4dc")
	test := make([]*TuxedoPopsTX.Ingredient, 1)
	test[0] = new(TuxedoPopsTX.Ingredient)
	test[0].Denominator = 1
	test[0].Numerator = 1
	test[0].Type = "Test Asset"

	recipeArgs.Ingredients = test

	sigHex := generateRecipeSig(recipeArgs.RecipeName, recipeArgs.CreatedType,
		recipeArgs.Ingredients, "94d7fe7308a452fdf019a0424d9c48ba9b66bdbca565c6fa3b1bf9c646ebac20")

	var err error
	recipeArgs.CreatorSig, err = hex.DecodeString(sigHex)
	if err != nil {
		HandleError(t, fmt.Errorf("error decoding creator signature in register recipe. ERR: (%v)", err.Error()))
		t.FailNow()
	}
	recipeArgsBytes, err := proto.Marshal(&recipeArgs)
	if err != nil {
		HandleError(t, fmt.Errorf("error marshalling recipeArgs in registerRecipe. ERR: (%s)\n", err.Error()))
		t.FailNow()
	}
	recipeArgsBytesStr := hex.EncodeToString(recipeArgsBytes)
	_, err = stub.MockInvoke("4", "recipe", []string{recipeArgsBytesStr})
	if err != nil {
		HandleError(t, err)
		t.Errorf("error invoking recipe: (%v)", err.Error())
	}
}

func generateCombineSig(counter string, combine TuxedoPopsTX.Combine, amount int, data string,
	privateKeyStr string) []byte {

	privKeyByte, _ := hex.DecodeString(privateKeyStr)
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyByte)
	message := counter
	message += ":" + combine.Recipe
	for _, source := range combine.GetSources() {
		message += ":" + strconv.FormatInt(int64(source.Idx()), 10)
		message += ":" + strconv.FormatInt(int64(source.Amount()), 10)
	}
	message += ":" + strconv.FormatInt(int64(amount), 10) + ":" + data
	fmt.Printf("\n\ncombine message: (%s)\n\n", message)
	messageBytes := sha256.Sum256([]byte(message))
	fmt.Println(message)

	sig, _ := privKey.Sign(messageBytes[:])

	return sig.Serialize()
}

func combine(t *testing.T, stub *shim.MockStub, popcode *keyInfo, sources []*TuxedoPopsTX.CombineSources,
	amount int32, recipe string, creator *keyInfo, owners []*keyInfo, data string) {

	var err error
	combineArgs := TuxedoPopsTX.Combine{}
	combineArgs.Address = popcode.address
	combineArgs.Sources = sources
	combineArgs.Amount = amount
	combineArgs.Recipe = recipe
	combineArgs.Data = data

	combineArgs.CreatorPubKey, err = hex.DecodeString(creator.pubKeyStr)
	if err != nil {
		HandleError(t, fmt.Errorf("error decoding public key: %v", err.Error()))
		t.FailNow()
	}

	popcode.counter, err = getCounter(stub, popcode)
	if err != nil {
		t.Errorf("error retrieving counterseed: (%v)", err.Error())
		t.FailNow()
	}
	combineArgs.CreatorSig = generateCombineSig(popcode.counter, combineArgs,
		int(combineArgs.Amount), combineArgs.Data, creator.privKeyStr)

	combineArgs.OwnerSigs = make([][]byte, 0)

	for _, owner := range owners {
		if owner != nil {
			ownerSig := generateCombineSig(popcode.counter, combineArgs,
				int(combineArgs.Amount), combineArgs.Data, owner.privKeyStr)

			combineArgs.OwnerSigs = append(combineArgs.OwnerSigs, ownerSig)
		}
	}
	combineArgs.PopcodePubKey, err = hex.DecodeString(popcode.pubKeyStr)
	if err != nil {
		HandleError(t, err)
		t.FailNow()
	}
	combineArgs.PopcodeSig = generateCombineSig(popcode.counter, combineArgs,
		int(combineArgs.Amount), combineArgs.Data, popcode.privKeyStr)

	combineArgsBytes, _ := proto.Marshal(&combineArgs)
	combineArgsBytesStr := hex.EncodeToString(combineArgsBytes)

	_, err = stub.MockInvoke("4", "combine", []string{combineArgsBytesStr})
	if err != nil {
		HandleError(t, fmt.Errorf("\nError invoking combine in checkCombine. ERR: (%s)", err.Error()))
		t.FailNow()
	}
}

func checkCombine(t *testing.T, stub *shim.MockStub) {
	registerRecipe(t, stub)
	function := "recipe"
	bytes, err := stub.MockQuery(function, []string{"test recipe"})
	if err != nil {
		HandleError(t, fmt.Errorf("Query (%s) failed. ERR: %v", function, err.Error()))
		t.FailNow()
	}
	if bytes == nil {
		HandleError(t, fmt.Errorf("Query (%s) failed to get value\n", function))
		t.FailNow()
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(bytes, &jsonMap); err != nil {
		HandleError(t, fmt.Errorf("error unmarshalling json string %s", bytes))
		t.FailNow()
	}
	fmt.Printf("JSON: %s\n", jsonMap)

	//create a new set of keys
	keys := new(keyInfo)
	keys.privKeyStr, err = newPrivateKeyString()
	if err != nil {
		HandleError(t, fmt.Errorf("error generating private key: %v", err))
	}
	keys.pubKeyStr, err = newPubKeyString(keys.privKeyStr)
	if err != nil {
		HandleError(t, fmt.Errorf("error generating public key: %v", err))
	}
	keys.address = newAddress(keys.pubKeyStr)
	keys.counter, err = getCounter(stub, keys)
	if err != nil {
		HandleError(t, fmt.Errorf("error retrieving counterseed: (%v)", err.Error()))
	}

	//mint value in popcode through create transaction with keys and counterseed
	altMint(t, stub, keys)

	keys.counter, err = getCounter(stub, keys)
	if err != nil {
		HandleError(t, fmt.Errorf("error retrieving counterseed: (%v)", err.Error()))
		t.FailNow()
	}

	//perform combination
	combineArgs := TuxedoPopsTX.Combine{}
	combineArgs.Address = keys.address
	//Sources
	source := new(TuxedoPopsTX.CombineSources)
	source.SourceAmount = 10
	source.SourceOutput = 0
	combineArgs.Sources = append(combineArgs.Sources, source)

	combineArgs.Amount = 10
	combineArgs.Recipe = "test recipe"
	combineArgs.Data = "test data"

	creatorPrivKey, _ := newPrivateKeyString()
	creatorPubKeyStr, _ := newPubKeyString(creatorPrivKey)

	combineArgs.CreatorPubKey, _ = hex.DecodeString(creatorPubKeyStr)
	if err != nil {
		HandleError(t, fmt.Errorf("generating private key: %v", err.Error()))
		t.FailNow()
	}

	combineArgs.CreatorSig = generateCombineSig(keys.counter, combineArgs,
		int(combineArgs.Amount), combineArgs.Data, creatorPrivKey)

	combineArgs.OwnerSigs = make([][]byte, 0)
	combineArgs.PopcodePubKey, _ = hex.DecodeString(keys.pubKeyStr)
	combineArgs.PopcodeSig = generateCombineSig(keys.counter, combineArgs,
		int(combineArgs.Amount), combineArgs.Data, keys.privKeyStr)

	combineArgsBytes, _ := proto.Marshal(&combineArgs)
	combineArgsBytesStr := hex.EncodeToString(combineArgsBytes)

	_, err = stub.MockInvoke("4", "combine", []string{combineArgsBytesStr})
	if err != nil {
		HandleError(t, fmt.Errorf("\nError invoking combine in checkCombine. ERR: (%s)", err.Error()))
		t.FailNow()
	}
}

/*
	//To create new private and public keys
	privKeyString, err := newPrivateKeyString()
	if err != nil {
		fmt.Println(err)
	}
	pubKeyString, err := newPubKeyString(privKeyString)
	if err != nil {
		fmt.Println(err)
	}
*/
//generates and returns SHA256 private key string
func newPrivateKeyString() (string, error) {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", fmt.Errorf("Error generating private key\n")
	}
	privKeyBytes := privKey.Serialize()
	privKeyString := hex.EncodeToString(privKeyBytes)
	return privKeyString, nil
}

//generates and returns SHA256 public key string fromessage private key string input
func newPubKeyString(privKeyString string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privKeyString)
	if err != nil {
		return "", fmt.Errorf("error decoding private key string (%s)", privKeyString)
	}
	_, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	pubKeyBytes := pubKey.SerializeCompressed()
	pubkKeyString := hex.EncodeToString(pubKeyBytes)
	return pubkKeyString, nil
}

//generates and returns first forty characters of sha256 hash of public key string
func newAddress(pubKeyStr string) string {
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		fmt.Printf("error decoding pubkeystring (%s)", pubKeyStr)
	}
	hasher := sha256.New()
	hasher.Write(pubKeyBytes)
	hashedPubKeyBytes := []byte{}
	hashedPubKeyBytes = hasher.Sum(hashedPubKeyBytes)
	hashedPubKeyString := hex.EncodeToString(hashedPubKeyBytes[0:20])
	address := hashedPubKeyString
	return address
}

func generateKeys() (*keyInfo, error) {
	var err error
	keys := new(keyInfo)
	keys.privKeyStr, err = newPrivateKeyString()
	if err != nil {
		fmt.Printf("error generating private key: %v", err.Error())
		return nil, fmt.Errorf("error generating private key: %v", err.Error())
	}
	keys.pubKeyStr, err = newPubKeyString(keys.privKeyStr)
	if err != nil {
		fmt.Printf("error generating public key: %v", err.Error())
		return nil, fmt.Errorf("error generating public key: %v", err.Error())
	}
	keys.address = newAddress(keys.pubKeyStr)
	return keys, nil
}

func getCounter(stub *shim.MockStub, keys *keyInfo) (string, error) {
	//query balance to get counterSeed
	bytes, err := stub.MockQuery("balance", []string{keys.address})
	if err != nil {
		return "", fmt.Errorf("balance query failure on address: (%s)\n", keys.address)
	}
	var balanceResult map[string]interface{}
	json.Unmarshal(bytes, &balanceResult)
	return balanceResult["Counter"].(string), nil
}

type balanceJSON struct {
	Address string
	Counter string
	Outputs []string
}

type finalBalanceJSON struct {
	Address string
	Counter string
	Outputs []outputJSON
}

type outputJSON struct {
	Owners      []string `json:"Owners"`
	Threshold   int64    `json:"Threshold"`
	Amount      int64    `json:"Amount"`
	Type        string   `json:"Type"`
	Data        string   `json:"Data"`
	Recipe      string   `json:"Recipe"`
	Creator     string   `json:"Creator"`
	PrevCounter string   `json:"PrevCounter"`
}

func getBalance(t *testing.T, stub *shim.MockStub, keys *keyInfo) finalBalanceJSON {
	balance := balanceJSON{}
	bytes, err := stub.MockQuery("balance", []string{keys.address})
	if err != nil {
		HandleError(t, fmt.Errorf("balance query failure on address: (%s)\n", keys.address))
	}

	err = json.Unmarshal(bytes, &balance)
	if err != nil {
		HandleError(t, fmt.Errorf("error unmarshalling balance for address :(%s)", keys.address))
		t.FailNow()
	}
	balanceResult := finalBalanceJSON{}
	balanceResult.Address = balance.Address
	balanceResult.Counter = balance.Counter
	for _, output := range balance.Outputs {
		var tmp outputJSON
		if err := json.Unmarshal([]byte(output), &tmp); err != nil {
			HandleError(t, fmt.Errorf("Error unmarshalling balance output (%v) into outputJSON struct\nERR: (%s)",
				output, err.Error()))
		}
		balanceResult.Outputs = append(balanceResult.Outputs, tmp)
	}
	return balanceResult
}

/*
	checkCounterSeedChange creates 150 popcodes and checks that the counterseed changes at the appropriate time.
	Can be called in a loop.
*/
func checkCounterSeedChange(t *testing.T, stub *shim.MockStub) {
	originalCounterseed, err := stub.GetState("CounterSeed")
	if err != nil {
		HandleError(t, fmt.Errorf("error retrieving counterseed through counterseed query"))
		t.FailNow()
	}
	txCache := txcache.TXCache{}
	txCacheBytes, err := stub.GetState("TxCache")
	if err != nil {
		HandleError(t, err)
		t.FailNow()
	}
	proto.Unmarshal(txCacheBytes, &txCache)
	//create up to 150 popcodes
	for i := len(txCache.Cache); i < 150; i++ {
		//create a new set of keys
		keys, err := generateKeys()
		if err != nil {
			HandleError(t, fmt.Errorf("error generating keys: (%v)\n", err.Error()))
		}

		keys.counter, err = getCounter(stub, keys)
		if err != nil {
			HandleError(t, fmt.Errorf("error retrieving counterseed: (%v)", err.Error()))
			t.FailNow()
		}

		//mint transaction with keys and counterseed
		altMint(t, stub, keys)

		//check counterseed
		counterseed, err := stub.GetState("CounterSeed")
		if err != nil {
			HandleError(t, fmt.Errorf("error retrieving counterseed through call to getState"))
		}

		txCache := txcache.TXCache{}
		txCacheBytes, err := stub.GetState("TxCache")
		if err != nil {
			HandleError(t, err)
		}
		proto.Unmarshal(txCacheBytes, &txCache)

		fmt.Printf("\n\nCOUNTERSEEDSTRING: (%s)\ni: (%d)\nTXCACHELEN: (%d)\n\n\n",
			hex.EncodeToString(counterseed), i, len(txCache.Cache))

		//check for correct counterSeed value
		if (i < 101) && (hex.EncodeToString(counterseed) != hex.EncodeToString(originalCounterseed)) {
			HandleError(t, fmt.Errorf("\nCounterseed got:\n(%s)\nwant:\n(%s)\n",
				hex.EncodeToString(counterseed), hex.EncodeToString(originalCounterseed)))

			t.FailNow()
		}
		if expected := sha256.Sum256(originalCounterseed); i > 101 &&
			(hex.EncodeToString(counterseed) != hex.EncodeToString(expected[:])) {

			HandleError(t, fmt.Errorf("\nCounterseed got:\n(%s)\nwant:\n(%s)\n",
				hex.EncodeToString(counterseed), hex.EncodeToString(expected[:])))

			t.FailNow()
		}
	}
}

func generateUsers(stub *shim.MockStub) (*users, error) {
	users := new(users)
	var err error
	users.user1, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user1.counter, err = getCounter(stub, users.user1)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	users.user2, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user2.counter, err = getCounter(stub, users.user2)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	users.user3, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user3.counter, err = getCounter(stub, users.user3)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	users.user4, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user4.counter, err = getCounter(stub, users.user4)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	users.user5, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user5.counter, err = getCounter(stub, users.user5)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	users.user6, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user6.counter, err = getCounter(stub, users.user6)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	users.user7, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generateUsers: (%v)\n", err.Error())
	}
	users.user7.counter, err = getCounter(stub, users.user7)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	return users, nil
}

func generatePopcodes(stub *shim.MockStub) (*popcodes, error) {
	popcodes := new(popcodes)
	var err error
	popcodes.popcode1, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode1.counter, err = getCounter(stub, popcodes.popcode1)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	popcodes.popcode2, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode2.counter, err = getCounter(stub, popcodes.popcode2)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	popcodes.popcode3, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode3.counter, err = getCounter(stub, popcodes.popcode3)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	popcodes.popcode4, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode4.counter, err = getCounter(stub, popcodes.popcode4)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	popcodes.popcode5, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode5.counter, err = getCounter(stub, popcodes.popcode5)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	popcodes.popcode6, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode6.counter, err = getCounter(stub, popcodes.popcode6)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	popcodes.popcode7, err = generateKeys()
	if err != nil {
		return nil, fmt.Errorf("error generating keys in generatePopcodes: (%v)\n", err.Error())
	}
	popcodes.popcode7.counter, err = getCounter(stub, popcodes.popcode7)
	if err != nil {
		return nil, fmt.Errorf("error generating counterSeed in generateUsers: (%v)\n", err.Error())
	}

	return popcodes, nil
}

type keyInfo struct {
	privKeyStr string
	pubKeyStr  string
	address    string
	counter    string
}

type test struct {
	t        *testing.T
	stub     *shim.MockStub
	users    *users
	popcodes *popcodes
}

type users struct {
	user1 *keyInfo
	user2 *keyInfo
	user3 *keyInfo
	user4 *keyInfo
	user5 *keyInfo
	user6 *keyInfo
	user7 *keyInfo
}

type popcodes struct {
	popcode1 *keyInfo
	popcode2 *keyInfo
	popcode3 *keyInfo
	popcode4 *keyInfo
	popcode5 *keyInfo
	popcode6 *keyInfo
	popcode7 *keyInfo
}

type possessInfo struct {
	t          *testing.T
	stub       *shim.MockStub
	popcode    *keyInfo
	prevOwners []*keyInfo
	newOwners  []*keyInfo
	idx        int
	data       string
}

func HandleError(t *testing.T, err error) (b bool) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		re := regexp.MustCompile("[^/]+$")
		t.Errorf("\x1b[32m\n[ERROR] in %s\tat line: %d\n%v\x1b[0m\n\n", re.FindAllString(fn, -1)[0], line, err)
		b = true
	}
	return
}

func TestPopcodeChaincode(t *testing.T) {
	bst := new(tuxedoPopsChaincode)
	stub := shim.NewMockStub("tuxedoPops", bst)
	checkInit(t, stub, []string{"Hello World"})

	checkQuery(t, stub, "74ded2036e988fc56e3cff77a40c58239591e921", `{"Address":"74ded2036e988fc56e3cff77a40c58239591e921","Counter":"af5eef44907ccdcc33051d035f32f42de0d093fac2fd9d15923448f6af46bc43","Outputs":null}`)
	mint(t, stub, "af5eef44907ccdcc33051d035f32f42de0d093fac2fd9d15923448f6af46bc43")
	checkQuery(t, stub, "74ded2036e988fc56e3cff77a40c58239591e921", `{"Address":"74ded2036e988fc56e3cff77a40c58239591e921","Counter":"1adb7c0c1b464fb45860355bf8e711312c608d01202197e58116a424f74af254","Outputs":["{\"Owners\":null,\"Threshold\":0,\"Data\":\"Test Data\",\"Type\":\"Test Asset\",\"PrevCounter\":\"e91d1eab53d597e8e18bb9ebbbaec66d08187d7e14a4a58c8782610ce7c7a74b\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}"]}`)
	mint(t, stub, "1adb7c0c1b464fb45860355bf8e711312c608d01202197e58116a424f74af254")
	checkQuery(t, stub, "74ded2036e988fc56e3cff77a40c58239591e921", `{"Address":"74ded2036e988fc56e3cff77a40c58239591e921","Counter":"afab4e267a433fe306d1da4608629ce9a280bde98f7004ff883383d65b9f5948","Outputs":["{\"Owners\":null,\"Threshold\":0,\"Data\":\"Test Data\",\"Type\":\"Test Asset\",\"PrevCounter\":\"e91d1eab53d597e8e18bb9ebbbaec66d08187d7e14a4a58c8782610ce7c7a74b\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}","{\"Owners\":null,\"Threshold\":0,\"Data\":\"Test Data\",\"Type\":\"Test Asset\",\"PrevCounter\":\"d3e41e748a7094cc520319623479f97dfb6aae0ea915940b72926384fe8d0e8c\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}"]}`)
	possess(t, stub, "afab4e267a433fe306d1da4608629ce9a280bde98f7004ff883383d65b9f5948", 1)
	checkQuery(t, stub, "74ded2036e988fc56e3cff77a40c58239591e921", `{"Address":"74ded2036e988fc56e3cff77a40c58239591e921","Counter":"92c7dff498fbe29d4b8d959a0f519a26ce43844f8871736191e5b62f8f507ea0","Outputs":["{\"Owners\":null,\"Threshold\":0,\"Data\":\"Test Data\",\"Type\":\"Test Asset\",\"PrevCounter\":\"e91d1eab53d597e8e18bb9ebbbaec66d08187d7e14a4a58c8782610ce7c7a74b\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}","{\"Owners\":[\"0278b76afbefb1e1185bc63ed1a17dd88634e0587491f03e9a8d2d25d9ab289ee7\"],\"Threshold\":1,\"Data\":\"Test possess\",\"Type\":\"Test Asset\",\"PrevCounter\":\"afab4e267a433fe306d1da4608629ce9a280bde98f7004ff883383d65b9f5948\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}"]}`)
	unitize(t, stub, "92c7dff498fbe29d4b8d959a0f519a26ce43844f8871736191e5b62f8f507ea0")
	checkQuery(t, stub, "74ded2036e988fc56e3cff77a40c58239591e921", `{"Address":"74ded2036e988fc56e3cff77a40c58239591e921","Counter":"92c7dff498fbe29d4b8d959a0f519a26ce43844f8871736191e5b62f8f507ea0","Outputs":["{\"Owners\":[\"0278b76afbefb1e1185bc63ed1a17dd88634e0587491f03e9a8d2d25d9ab289ee7\"],\"Threshold\":1,\"Data\":\"Test possess\",\"Type\":\"Test Asset\",\"PrevCounter\":\"afab4e267a433fe306d1da4608629ce9a280bde98f7004ff883383d65b9f5948\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}"]}`)
	checkQuery(t, stub, "10734390011641497f489cb475743b8e50d429bb", `{"Address":"10734390011641497f489cb475743b8e50d429bb","Counter":"3d2cc9f7d475cf79347ff317b1164daa50ced56d3ee977252da0430f39fa7a4e","Outputs":["{\"Owners\":null,\"Threshold\":0,\"Data\":\"Test Unitize\",\"Type\":\"Test Asset\",\"PrevCounter\":\"660bfdba4544847711d515fb26c5f1f62f0c9fc45b5a41b0fefcc1d58de4f1c0\",\"Creator\":\"03cc7d40833fdf46e05a7f86a6c9cf8a697a129fbae0676ad6bad71f163ea22b26\",\"Amount\":10}"]}`)

	for i := 0; i < 2; i++ {
		checkCounterSeedChange(t, stub)
	}
	checkCombine(t, stub)

	//new testing suite below:
	fmt.Printf("\n\n\nSTARTING NEW TESTING SUITE\n\n\n")

	users, err := generateUsers(stub)
	if err != nil {
		HandleError(t, fmt.Errorf("error generating users: (%v)\n", err.Error()))
		t.FailNow()
	}

	popcodes, err := generatePopcodes(stub)
	if err != nil {
		HandleError(t, fmt.Errorf("error generating popcodes: (%v)\n", err.Error()))
		t.FailNow()
	}

	//MINT
	balance := getBalance(t, stub, popcodes.popcode1)
	fmt.Printf("\n\n\nbalance before mint (%v)\n\n\n", balance)
	prevCounter := balance.Counter
	prevNumberOfOutputs := len(balance.Outputs)
	data := "data"
	createdType := "Water"
	amount := 100

	altMint1(t, stub, users.user1, popcodes.popcode1, data, createdType, amount)
	balance = getBalance(t, stub, popcodes.popcode1)
	fmt.Printf("\n\n\nbalance on popcode (%v)\ncounter: (%v)\noutputs: (%v)\n\n",
		balance.Address, balance.Counter, balance.Outputs)

	if prevCounter == balance.Counter {
		HandleError(t, fmt.Errorf("counter of address (%s) did not change after call to mint. Counter: (%s)",
			balance.Address, balance.Counter))
		t.FailNow()
	}
	if len(balance.Outputs) != prevNumberOfOutputs+1 {
		HandleError(t, fmt.Errorf("number of outputs of popcode with address (%s)"+
			" did not increase by one after create transaction", popcodes.popcode1))
		t.FailNow()
	}

	/*
		POSSESS
		perform initial possess
		one new owner
	*/
	output := 0
	prevCounter = balance.Counter
	prevOwners := make([]*keyInfo, 1)
	newOwners := make([]*keyInfo, 0)
	newOwners = append(newOwners, users.user1)
	data = "data"
	threshold := len(newOwners)
	altPossess(t, stub, popcodes.popcode1, prevOwners, newOwners, output, data, threshold)
	balance = getBalance(t, stub, popcodes.popcode1)
	if prevCounter == balance.Counter {
		HandleError(t, fmt.Errorf("counter of address (%s) did not change after call to possess. Counter: (%s)",
			balance.Address, balance.Counter))
	}
	if len(balance.Outputs[output].Owners) != len(newOwners) {
		HandleError(t, fmt.Errorf("after possess on unowned popcode with address: (%s)\n"+
			"number of owners got (%d). Want (%d)\noutputs: (%v)",
			balance.Address, len(balance.Outputs[output].Owners), len(newOwners), balance.Outputs))
	}
	for i, owner := range newOwners {
		if balance.Outputs[output].Owners[i] != owner.pubKeyStr {
			HandleError(t, fmt.Errorf("incorrect owner at index (%d). Got (%s). Want (%s)\n",
				i, balance.Outputs[output].Owners[i], owner.pubKeyStr))
		}
	}

	/*
		possess an owned popcode
		multiple new owners
		check counterseed change, number of owners, owner change
		TODO check threshold functionality
	*/
	prevCounter = balance.Counter
	prevOwners[0] = newOwners[0]
	newOwners = []*keyInfo{users.user1, users.user2, users.user3, users.user4, users.user5, users.user6}
	fmt.Printf("\n\n\nprevOwners: (%v)\nnewOwners: (%v)\n\n\n\n", prevOwners[0], newOwners[0])
	data = "data"
	output = 0
	threshold = len(newOwners)
	altPossess(t, stub, popcodes.popcode1, prevOwners, newOwners, output, data, threshold)
	balance = getBalance(t, stub, popcodes.popcode1)
	if prevCounter == balance.Counter {
		HandleError(t, fmt.Errorf("counter of address (%s) did not change after call to possess. Counter: (%s)",
			balance.Address, balance.Counter))
	}
	if len(balance.Outputs[output].Owners) != len(newOwners) {
		HandleError(t, fmt.Errorf("after possess on unowned popcode with address: (%s)\n"+
			"number of owners got (%d). Want (%d)\noutputs: (%v)",
			balance.Address, len(balance.Outputs[output].Owners), len(newOwners), balance.Outputs))
	}
	for i, owner := range newOwners {
		if balance.Outputs[output].Owners[i] != owner.pubKeyStr {
			HandleError(t, fmt.Errorf("incorrect owner at index (%d). Got (%s). Want (%s)\n",
				i, balance.Outputs[output].Owners[i], owner.pubKeyStr))
		}
	}

	/*
		possess and change threshold to 1
		check threshold change
	*/
	prevCounter = balance.Counter
	owners := newOwners
	newOwners = []*keyInfo{users.user1, users.user2}
	data = "data"
	output = 0
	threshold = 1
	altPossess(t, stub, popcodes.popcode1, owners, newOwners, output, data, threshold)
	balance = getBalance(t, stub, popcodes.popcode1)
	if prevCounter == balance.Counter {
		HandleError(t, fmt.Errorf("counter of address (%s) did not change after call to possess. Counter: (%s)",
			balance.Address, balance.Counter))
	}
	if len(balance.Outputs[output].Owners) != len(newOwners) {
		HandleError(t, fmt.Errorf("after possess on unowned popcode with address: (%s)\n"+
			"number of owners got (%d). Want (%d)\noutputs: (%v)",
			balance.Address, len(balance.Outputs[output].Owners), len(newOwners), balance.Outputs))
	}
	for i, owner := range newOwners {
		if balance.Outputs[output].Owners[i] != owner.pubKeyStr {
			HandleError(t, fmt.Errorf("incorrect owner at index (%d). Got (%s). Want (%s)\n",
				i, balance.Outputs[output].Owners[i], owner.pubKeyStr))
		}
	}
	if balance.Outputs[output].Threshold != int64(threshold) {
		HandleError(t, fmt.Errorf("threshold did not change as expected\nGot (%d)\nWant (%d)", balance.Outputs[output].Threshold, threshold))
	}
	fmt.Printf("\n\nNew BALANCE: (%v)\n\n\n", balance)

	/*
		check threshold functionality
		possess using only one of two owner sigs
	*/
	prevCounter = balance.Counter
	prevOwners = []*keyInfo{users.user1}
	newOwners[0] = users.user2
	newOwners[1] = users.user3
	fmt.Printf("\n\n\nprevOwners: (%v)\nnewOwners: (%v)\n\n\n\n", prevOwners[0], newOwners[0])
	data = "data"
	output = 0
	threshold = len(newOwners)
	altPossess(t, stub, popcodes.popcode1, prevOwners, newOwners, output, data, threshold)
	balance = getBalance(t, stub, popcodes.popcode1)
	if prevCounter == balance.Counter {
		HandleError(t, fmt.Errorf("counter of address (%s) did not change after call to possess. Counter: (%s)",
			balance.Address, balance.Counter))
	}
	if len(balance.Outputs[output].Owners) != len(newOwners) {
		HandleError(t, fmt.Errorf("after possess on unowned popcode with address: (%s)\n"+
			"number of owners got (%d). Want (%d)\noutputs: (%v)",
			balance.Address, len(balance.Outputs[output].Owners), len(newOwners), balance.Outputs))
	}
	for i, owner := range newOwners {
		if balance.Outputs[output].Owners[i] != owner.pubKeyStr {
			HandleError(t, fmt.Errorf("incorrect owner at index (%d). Got (%s). Want (%s)\n",
				i, balance.Outputs[output].Owners[i], owner.pubKeyStr))
		}
	}

	owners = newOwners
	testUnitize(t, stub, popcodes, users, owners)

	/*
		RECIPE
	*/
	ingredients := []*TuxedoPopsTX.Ingredient{}
	ingredient := new(TuxedoPopsTX.Ingredient)
	ingredient.Denominator = 1
	ingredient.Numerator = 1
	/*
		TODO:
		figure out why it doesn't matter that ingredient.Type is " ".
		Why does the combination still work?
	*/
	ingredient.Type = " "
	recipeName := "Water Vapor Recipe"
	createdType = "Water Vapor"
	altRecipe(t, stub, recipeName, createdType, users.user1, ingredients)

	testCombine(t, stub, popcodes, users, recipeName, owners)
}

/*
	UNITIZE
	check the following:
		counterseed change,
		change in number of outputs, and
		change in quantity of units in ouputs
*/
func testUnitize(t *testing.T, stub *shim.MockStub, popcodes *popcodes, users *users, owners []*keyInfo) {
	sourceBalance := getBalance(t, stub, popcodes.popcode1)
	//unitize owned output into two outputs in a different popcode
	fmt.Printf("\n\n\nbefore unitize: balance on source popcode (%v)\ncounter: (%v)\noutputs: (%v)\n\n\n",
		sourceBalance.Address, sourceBalance.Counter, sourceBalance.Outputs)
	// prevNumberOfOutputs := len(balance["Outputs"].([]interface{}))
	sourcePrevCounter := sourceBalance.Counter
	destBalance := getBalance(t, stub, popcodes.popcode2)
	fmt.Printf("\n\n\nbefore unitize: balance on destination popcode (%v)\ncounter: (%v)\noutputs: (%v)\n\n\n",
		destBalance.Address, destBalance.Counter, destBalance.Outputs)

	destPrevCounter := destBalance.Counter
	destAmounts := []int32{50, 50}
	data := "data"
	output := 0
	altUnitize(t, stub, popcodes.popcode1, popcodes.popcode2, owners, data, destAmounts, int32(output))
	sourceBalance = getBalance(t, stub, popcodes.popcode1)
	if sourcePrevCounter != sourceBalance.Counter {
		HandleError(t, fmt.Errorf("counter of source popcode (address: %s) changed after call to unitize. Counter: (%s)",
			sourceBalance.Address, sourceBalance.Counter))
	}
	destBalance = getBalance(t, stub, popcodes.popcode2)
	if destPrevCounter == destBalance.Counter {
		HandleError(t, fmt.Errorf("counter of destination popcode (address: %s) "+
			"did not change after call to unitize. Counter: (%s)", destBalance.Address, destBalance.Counter))
	}

	fmt.Printf("\n\n\nafter unitize: balance on source popcode (%v)\ncounter: (%v)\noutputs: (%v)\n\n",
		sourceBalance.Address, sourceBalance.Counter, sourceBalance.Outputs)
	fmt.Printf("\n\n\nafter unitize: balance on destination popcode (%v)\ncounter: (%v)\noutputs: (%v)\n\n",
		destBalance.Address, destBalance.Counter, destBalance.Outputs)

}

/*
	COMBINE
	combination transaction with two owners
*/
func testCombine(t *testing.T, stub *shim.MockStub, popcodes *popcodes, users *users, recipeName string, owners []*keyInfo) {
	popcode := popcodes.popcode2
	sources := []*TuxedoPopsTX.CombineSources{}
	source := new(TuxedoPopsTX.CombineSources)
	source.SourceOutput = 0
	source.SourceAmount = 1
	sources = append(sources, source)
	creator := users.user2
	data := "data"
	amount := 1
	// recipeName = "Water Vapor Recipe"

	popcodeBalance := getBalance(t, stub, popcodes.popcode2)
	fmt.Printf("\n\npopcode balance: (%v)\n\n", popcodeBalance)

	prevCounter := popcodeBalance.Counter
	prevNumberOfOutputs := len(popcodeBalance.Outputs)
	combine(t, stub, popcode, sources, int32(amount), recipeName, creator, owners, data)
	popcodeBalance = getBalance(t, stub, popcodes.popcode2)
	if popcodeBalance.Counter == prevCounter {
		HandleError(t, fmt.Errorf("Counter of popcode (%s) did not change after call to combine\n", popcode.address))
	}
	if len(popcodeBalance.Outputs) == prevNumberOfOutputs {
		HandleError(t, fmt.Errorf("Number of outputs of popcode (%d) did not change after call to combine"+
			"\npopcode balance: (%v)", len(popcodeBalance.Outputs), popcodeBalance))
	}
	fmt.Printf("popcode balance: (%v)\n\n", popcodeBalance)
}
