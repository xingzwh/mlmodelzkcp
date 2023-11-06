package main

import (
	"fmt"
	"log"
	"os/exec"
	"bytes"
	_"strconv"
	"io/ioutil"
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type CC_Seller struct {
}

func (t *CC_Seller) Init(stub shim.ChaincodeStubInterface) pb.Response {
	
	return shim.Success([]byte("successful init"))
}


func (t *CC_Seller) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()	
	if fn == "uploadvalue" {
		return t.UploadValue(stub, args)				
	}else if fn == "getvalue"{
		return t.GetValue(stub, args)		
	}else if fn == "reproduce"{
		return t.Reproduce(stub, args)		
	}else if fn == "setupproof"{
		return t.LRGroth16Setup(stub, args)		
	}else{
		return shim.Error("no such opration")
	}
	
}

func (t *CC_Seller) GetValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	

	var keyID = args[0]
	keyvalue,err:=stub.GetState(keyID)
	if err != nil {
		return shim.Error("error getvalue")
	}
	fmt.Println(keyID)
	fmt.Println(keyvalue)
	return shim.Success([]byte(keyvalue))
}
func (t *CC_Seller) UploadValue(stub shim.ChaincodeStubInterface, args []string) pb.Response{	
	var keyID = args[0]
	var keyvalue = args[1]
	stub.PutState(keyID,[]byte(keyvalue))
	fmt.Println(keyID)
	fmt.Println(keyvalue)
	return shim.Success([]byte("success put " + keyID + " " + keyvalue))
}



func (t *CC_Seller) Reproduce(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	keyID = args[0]
	pyUrl := args[1]      ///xx/xx/xx.py 	具体到文件,py文件路径
	dataUrl = args[2]    ///xx/xx/xx.csv	具体到文件，数据文件路径
	envUrl = args[3]    ///xx/xx/	文件所属文件夹，用于进行hash保存
	parameterUrl = args[4]    ///xx/newxx/  参数存储路径
	chaincodeID = args[5]
	
	//调用prover下的链玛复现训练流程
        reproduceArgs :=[][]byte{[]byte("reproduce"),[]byte(keyID),[]byte(pyUrl),[]byte(dataUrl),[]byte(envUrl),[]byte(parameterUrl)}
            reproduceResponse:=stub.InvokeChaincode(chaincodeID,reproduceArgs,"testch1")
            if reproduceResponse.Status != shim.OK {
            	return shim.Error("failed to query chaincode.got error :"+string(reproduceResponse.Payload))
            }            
	fmt.Println(reproduceResponse.Payload)
	return shim.Success([]byte(reproduceResponse.Payload))
	
}


func (t *CC_Seller) SetupProof(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	keyID = args[0]
	
	chaincodeId:= args[1]
	//具体到路径，由于保存参数的文件不一样，所以文件在Prover链玛中指定
	parameterUrl := args[2]
	datasetUrl := args[3]	//xx/xx/xx.py 	具体到文件
	dataIndex := args[4]	//默认为“0”，如果不为0则使用index的数值
	//指定阈值
	
	thrName:= keyID + "threshold"
	thrDate,err:=stub.GetState(thrName)
	if err != nil {
		return shim.Error("error getvalue")
	}
	
	
	queryArgs :=[][]byte{[]byte("setupproof"),[]byte(keyID),[]byte(dataUrl),[]byte(parameterUrl),[]byte(thrDate),[]byte(dataIndex)}
	response:=stub.InvokeChaincode(chaincodeId,queryArgs,"testch1")
	if response.Status != shim.OK {
		return shim.Error("failed to query chaincode.got error :"+string(response.Payload))
	}
	fmt.Printf(string(response.Payload))	

	return shim.Success([]byte("success setproof"))
}


func main() {
	if err := shim.Start(new(CC_Seller)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
