package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	//"crypto/x509"
	//"encoding/pem"
	//"net/url"
	"regexp"
	"time"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const   KAKACENTER   =  1 	//AUTHORITY
const   SHOP   =  2			//MANUFACTURER
const   CONSUMER =  3		//PRIVATE_ENTITY
const   MAILBOX  =  4		//LEASE_COMPANY
//const   SCRAP_MERCHANT =  5	

const	USER_HOLDER = "user_holder"
const	CARD_TEMPLATE_HOLDER = "card_template_holder"
const	CARD_HOLDER = "card_holder"


//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can 
//					be done to the card at points in it's lifecycle
//==============================================================================================================================
const   STATE_TEMPLATE  				=  0
const   STATE_SHOP  					=  1		//STATE_MANUFACTURE
const   STATE_CONSUMER_OWNERSHIP 		=  2		//STATE_PRIVATE_OWNERSHIP
const   STATE_MAILBOX_OWNERSHIP 		=  3		//STATE_LEASED_OUT
//const   STATE_BEING_SCRAPPED  			=  4

//==============================================================================================================================
//	 Structure Definitions 
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  CardTransactionChaincode struct {
}

//==============================================================================================================================
//	Card - Defines the structure for a car object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Card struct {
	Kakaid			string `json:"kakaid"`
	Shop			string `json:"shop"`
	Shopid			string `json:"shopid"`
	Cardid			string `json:"cardid"`
	Category		string `json:"category"`
	Cardlevel		string `json:"cardlevel"`
	Cardclass		string `json:"cardclass"`
	Owner			string `json:"owner"`
	Tel				string `json:"tel"`
	Password		string `json:"password"`
	Money			int `json:"money"`
	Point			int `json:"point"`
	Expdate			string `json:"expdate"`
	Getdate			string `json:"getdate"`
	Releasedate		string `json:"releasedate"`
	Expired			bool `json:"expired"`
	Scrapped       	bool `json:"scrapped"`
	Status       	int `json:"status"`
}


//==============================================================================================================================
//	Card_Holder - Defines the structure that holds all the Card for cards that have been created.
//				Used as an index when querying all cards.
//==============================================================================================================================

type Card_Holder struct {
	Cards 	[]string `json:"cards"`
}

//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================

type User struct {
	Identity 		string `json:"identity"`
	ECert 			string `json:"ecert"`
	Affiliation 	int `json:"affiliation"`
}	

type User_Holder struct {
	Users 		[]string `json:"users"`
}	

type ShopLedger struct {
	Templateid 		string `json:"templateid"`
	Shopid			string `json:"shopid"`
	CardIdIndex		int  `json:"cardIdIndex"`
	Qty 			int `json:"qty"`
	ExpiredNum		int `json:"expiredNum"`
	ScrapNum		int  `json:"scrapNum"`
	BackNum			int  `json:"backNum"`
	InitMoney 	int `json:"initmoney"`
	InitPoint 	int `json:"initpoint"`
	DepositMoney 	int `json:"depositMoney"`
	DepositPoint 	int `json:"tdepositPoint"`
	ConsumeMoney 	int `json:"consumeMoney"`
	ConsumePoint 	int `json:"consumePoint"`
}	

type ShopLedger_Holder struct {
	ShopLedgers 		[]string `json:"shopLedgers"`
}	

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode																	
//==============================================================================================================================
func (t *CardTransactionChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	// create and put card_holder
	var cardKakaIDs Card_Holder
	bytes, err := json.Marshal(cardKakaIDs)
	if err != nil { return nil, errors.New("Error creating Card_Holder record") }											
	err = stub.PutState(CARD_HOLDER, bytes)
	
	// create and put card_holder
	var user_holder User_Holder
	bytes, err = json.Marshal(user_holder)
	if err != nil { return nil, errors.New("Error creating User_Holder record") }														
	err = stub.PutState(USER_HOLDER, bytes)

	//add users
	for i:=0; i < len(args); i=i+3 {
		var user User
		user.Identity = args[i]
		user.ECert = args[i+1]
		user.Affiliation, err = strconv.Atoi(args[i+2])
			if err != nil { return nil, errors.New("ERROR: affiliation value is not int") }
		t.add_user(stub, user)													
	}


	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 add_user - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================
/*
func (t *CardTransactionChaincode) init_shopLedger_holder(stub shim.ChaincodeStubInterface, shop string) (ShopLedger_Holder, error) {
	
	shopLedgersBytes, err := stub.GetState(shop)
	if err != nil {	fmt.Printf("RETRIEVE_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("RETRIEVE_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}

	var shopLedger_holder ShopLedger_Holder
	if shopLedgersBytes == nil || len(shopLedgersBytes) == 0 {
		
		bytes, err := json.Marshal(shopLedger_holder)
		if err != nil { return nil, errors.New("Error creating User bytes") }
	

		err = stub.PutState(shop, bytes)
		if err != nil { fmt.Printf("SAVE_CHANGES: Error storing  card record: %s", err); return false, errors.New("Error storing card record") }
	}else {
		err = json.Unmarshal(shopLedgersBytes, &shopLedger_holder);						
		if err != nil {	fmt.Printf("Unmarshal_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("Unmarshal_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	}

	return shopLedger_holder, nil
}
*/

func (t *CardTransactionChaincode) get_shopLedgerID(shop string, templateID string) (string) {
	return templateID + "-" + shop
}

func (t *CardTransactionChaincode) add_new_shopLedger(stub shim.ChaincodeStubInterface, shopid string, templateID string) ([]byte, error) {
	
	shopLedgerId := t.get_shopLedgerID(shopid, templateID)
	shopLedgerBytes, err := stub.GetState(shopLedgerId)
	if err != nil {	fmt.Printf("add_new_shopLedger ERROR: %s", err); 
					return nil, errors.New("add_new_shopLedger ERROR")	}
	if shopLedgerBytes != nil || len(shopLedgerBytes) > 0 {				
		if err != nil {	fmt.Printf("shop ledger already exist, error: %s", err); 
					return nil, errors.New("shop ledger already exist")	}
	}
	
	var shopLedger ShopLedger																																									

	Templateid          	:= "\"Templateid\":\""+templateID+"\", "	
	Shopid          		:= "\"Shopid\":\""+shopid+"\", "
	CardIdIndex          	:= "\"CardIdIndex\":0, "	
	Qty          			:= "\"Qty\":0, "							
	ExpiredNum            	:= "\"ExpiredNum\":0, "
	ScrapNum        		:= "\"ScrapNum\":0, "
	BackNum       			:= "\"BackNum\":0, "
	InitMoney       		:= "\"InitMoney\":0, "
	InitPoint           	:= "\"InitPoint\":0, "
	DepositMoney       		:= "\"DepositMoney\":0, "
	TotalDepositPoint  	    := "\"TotalDepositPoint\":0, "
	ConsumeMoney        	:= "\"ConsumeMoney\":0, "
	ConsumePoint           	:= "\"ConsumePoint\":0 "
	

	shopLedger_json := "{"+Templateid+Shopid+CardIdIndex+Qty+ExpiredNum+ScrapNum+BackNum+InitMoney+InitPoint+DepositMoney+TotalDepositPoint+ConsumeMoney+ConsumePoint+"}" 	// Concatenates the variables to create the total JSON object
	
	fmt.Printf("shopLedger_json : %s ",shopLedger_json);

	err = json.Unmarshal([]byte(shopLedger_json), &shopLedger)	

	err = stub.PutState(shopLedgerId, []byte(shopLedger_json))
		if err != nil { fmt.Printf("SAVE_CHANGES: Error storing  card record: %s", err); return nil, errors.New("Error storing card record") }
	
	return []byte(shopLedger_json), nil
}


func (t *CardTransactionChaincode) get_shopLedger(stub shim.ChaincodeStubInterface, shopid string, templateID string) ([]byte, error) {
	
	shopLedgerId := t.get_shopLedgerID(shopid, templateID)
	shopLedgerBytes, err := stub.GetState(shopLedgerId)
	if err != nil {	fmt.Printf("get_shopLedger ERROR: %s", err); 
					return nil, errors.New("get_shopLedger ERROR")	}
																																							
	return shopLedgerBytes, nil
}

func (t *CardTransactionChaincode) update_shopLedger(stub shim.ChaincodeStubInterface, shopid string, templateID string, shopLedger ShopLedger) ([]byte, error) {
	
	shopLedgerId := t.get_shopLedgerID(shopid, templateID)
																															
	shopLedgersBytes, err := json.Marshal(shopLedger)	
	if err != nil {	fmt.Printf("Unmarshal shopLedgersBytes error: %s", err); 
					return nil, errors.New("Unmarshal shopLedgersBytes error")	}

	err = stub.PutState(shopLedgerId, shopLedgersBytes)
		if err != nil { fmt.Printf("PutState shopLedgersBytes: Error : %s", err); 
						return nil, errors.New("PutState shopLedgersBytes: Error") }
	
	return shopLedgersBytes, nil
}

/*
func (t *CardTransactionChaincode) get_shopLedger_holder(stub shim.ChaincodeStubInterface, shop string) (ShopLedger_Holder, error) {
	
	shopLedgersBytes, err := stub.GetState(shop)
	if err != nil {	fmt.Printf("RETRIEVE_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("RETRIEVE_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}

	var shopLedger_holder ShopLedger_Holder
	if shopLedgersBytes != nil || len(shopLedgersBytes) > 0 {
		err = json.Unmarshal(shopLedgersBytes, &shopLedger_holder);						
		if err != nil {	fmt.Printf("Unmarshal_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("Unmarshal_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	
	}

	return shopLedger_holder, nil
}
*/

func (t *CardTransactionChaincode) add_user(stub shim.ChaincodeStubInterface, user User) ([]byte, error) {
	

	ubytes, err := json.Marshal(user)
	if err != nil { return nil, errors.New("Error creating User bytes") }
	
	fmt.Printf("------------add user - new user bytes: "+string(ubytes)); 

	var user_holder User_Holder
	usersBytes, err := stub.GetState(USER_HOLDER)
	if err != nil {	fmt.Printf("RETRIEVE_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("RETRIEVE_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	fmt.Printf("------------add user - user_holder bytes: "+string(usersBytes)); 

	err = json.Unmarshal(usersBytes, &user_holder);						
	if err != nil {	fmt.Printf("Unmarshal_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("Unmarshal_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	
	fmt.Printf("------------add user - user_holder users num: " + string(len(user_holder.Users))); 

	user_holder.Users = append(user_holder.Users, string(ubytes))

    fmt.Printf("------------add user - new user_holder users num: " + string(len(user_holder.Users)));

	usersBytes, err = json.Marshal(user_holder)
	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting card card record: %s", err); return nil, errors.New("Error converting card record") }
	
	fmt.Printf("------------add user - new user_holder bytes: "+string(usersBytes)); 

	err = stub.PutState(USER_HOLDER, usersBytes)


	if err != nil {
		fmt.Printf("------------put states error: %s", err)
		fmt.Printf("------------put states error,user iendtiry: "+ user.Identity)
		return nil, errors.New("Error storing user: " + user.Identity )
	}
	
	fmt.Printf("------------put states success,user iendtiry: "+ user.Identity)

	return nil, nil

}

//==============================================================================================================================
//	 check_affiliation
//==============================================================================================================================

func (t *CardTransactionChaincode) check_affiliation(stub shim.ChaincodeStubInterface, userName string) (int, error) {
	
	user, err := t.get_user_detail_Internal(stub, userName)
	if err != nil { return -1, err }
	
	return user.Affiliation, nil
}


//==============================================================================================================================
//	 get_users - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *CardTransactionChaincode) get_users(stub shim.ChaincodeStubInterface,caller string) ([]byte, error) {

	//caller_affiliation, _ := t.check_affiliation(stub, caller)

	//if caller_affiliation == KAKACENTER {
		usersBytes, err := stub.GetState(USER_HOLDER)
		if err != nil {	fmt.Printf("RETRIEVE_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return usersBytes, errors.New("RETRIEVE_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	
		fmt.Printf("get_users , User_Holder bytes: "+string(usersBytes)); 

		return usersBytes, nil
	//}

	//return nil, errors.New("Permission denied: you are not KAKACENTER users ")
	
}


//==============================================================================================================================
//	 get_user_detail - 
//==============================================================================================================================
func (t *CardTransactionChaincode) get_user_detail_Internal(stub shim.ChaincodeStubInterface, name string) (User, error) {
	
	var u User

	var user_holder User_Holder
	usersBytes, err := stub.GetState(USER_HOLDER)
	if err != nil {	fmt.Printf("RETRIEVE_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return u, errors.New("RETRIEVE_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	
	fmt.Printf("User_Holder bytes: "+string(usersBytes)); 

	err = json.Unmarshal(usersBytes, &user_holder);						
	if err != nil {	fmt.Printf("Unmarshal_USER_HOLDER ERROR: Corrupt users record "+string(usersBytes)+": %s", err); 
					return u, errors.New("Unmarshal_USER_HOLDER ERROR: Corrupt users record"+string(usersBytes))	}
	
	
	
	for _, userStr := range user_holder.Users {
		

		err = json.Unmarshal([]byte(userStr), &u);						
		if err != nil {	fmt.Printf("Unmarshal_userStr: Corrupt user record "+userStr+": %s", err); 
		return u, errors.New("Unmarshal_userStr: Corrupt user record"+userStr)	}
	
		if u.Identity == name {return u,nil}
		
	}
	if err != nil { return u, errors.New("Couldn't retrieve user (" + name + ") from user_holder ") }
	
	return u, nil
}


//==============================================================================================================================
//	 get_user_detail - 
//==============================================================================================================================
func (t *CardTransactionChaincode) get_user_detail(stub shim.ChaincodeStubInterface, caller string, name string) (User, error) {
	
	return t.get_user_detail_Internal(stub, name)
}



//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================
/*
func (t *CardTransactionChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

	bytes, err := stub.GetCallerCertificate();
															if err != nil { return "", errors.New("Couldn't retrieve caller certificate") }
	x509Cert, err := x509.ParseCertificate(bytes);				// Extract Certificate from result of GetCallerCertificate						
															if err != nil { return "", errors.New("Couldn't parse certificate")	}
															
	return x509Cert.Subject.CommonName, nil
}
*/
//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================
/*
func (t *CardTransactionChaincode) check_affiliation(stub shim.ChaincodeStubInterface, cert string) (int, error) {																																																					
	

	decodedCert, err := url.QueryUnescape(cert);    				// make % etc normal //
	
		if err != nil { return -1, errors.New("Could not decode certificate") }
	
	pem, _ := pem.Decode([]byte(decodedCert))           				// Make Plain text   //

	x509Cert, err := x509.ParseCertificate(pem.Bytes);				// Extract Certificate from argument //
														
		if err != nil { return -1, errors.New("Couldn't parse certificate")	}

	cn := x509Cert.Subject.CommonName
	
	res := strings.Split(cn,"\\")
	
	affiliation, _ := strconv.Atoi(res[2])
	
	return affiliation, nil
		
}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *CardTransactionChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, int, error){	

	user, err := t.get_username(stub)
	if err != nil { return "", -1, err }
																		
	ecert, err := t.get_ecert(stub, user);					
	if err != nil { return "", -1, err }

	affiliation, err := t.check_affiliation(stub,string(ecert));			
	if err != nil { return "", -1, err }

	return user, affiliation, nil
}
*/



//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *CardTransactionChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	//caller, caller_affiliation, err := t.get_caller_data(stub)
	caller := args[0]
 	caller_affiliation , err := t.check_affiliation(stub, caller)
	if err != nil  { return nil, errors.New("Invalid caller_affiliation value passed") }
	
	cardIDPos := 1


	if function == "add_user" { 
		var user User
		user.Identity = args[0]
		user.ECert = args[1]
		user.Affiliation, err = strconv.Atoi(args[2])
		if err != nil { fmt.Printf("strconv.Atoi(args[2]) error: ", err); 
						return nil, errors.New("strconv.Atoi(args[2]) error") }
		return t.add_user(stub, user)

	} else if function == "create_card_template" { 
		fmt.Printf("------------create template function----------");
		return t.create_card_template(stub, caller, caller_affiliation, args[cardIDPos])

	} else if function == "transfer_template_to_shop" {

		cardTemplateId := args[1]
		cardTemplate, err := t.retrieve_card(stub, cardTemplateId)
		if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
						return nil, errors.New("Error retrieving v5c") }

		rec_user := args[2]
		rec_affiliation , err := t.check_affiliation(stub, rec_user)
			if err != nil  { return nil, errors.New("Invalid rec_affiliation value passed") }
		return t.transfer_template_to_shop(stub, cardTemplate, caller, caller_affiliation, rec_user, rec_affiliation)

	//} else if function == "update_template" {
	//	cardTemplateId := args[1]
	//	updatedCardJson := args[2]

	//	return t.update_template(stub, caller, caller_affiliation, cardTemplateId, updatedCardJson)

	} else if function == "create_card_by_template" { 
		cardTemplateId := args[1]
		cardNum, err := strconv.Atoi(args[2])
		if err != nil { fmt.Printf("strconv.Atoi(args[2]) card number error: ", err); 
						return nil, errors.New("strconv.Atoi(args[3]) card number error") }

		//create_card_by_template(stub , caller string, caller_affiliation int, 
		//							cardTemplate_KakaIDs string, initCard Card, cardIDPrefix string, cardNum int)
		return t.create_card_by_template(stub, caller, caller_affiliation, cardTemplateId, cardNum)

	} else if function == "scrap_card" {
		
		cardid := args[1]
		card, err := t.retrieve_card(stub, cardid)
		if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
						return nil, errors.New("Error retrieving v5c") }
		return t.scrap_card(stub, card, caller)

	} else if strings.Contains(function, "update") == true{

		cardid := args[1]
		newValue := args[2]
		card, err := t.retrieve_card(stub, cardid)
		if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
						return nil, errors.New("Error retrieving v5c") }

		if function == "update_shop"  	    { return t.update_shop(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_shopid"       { return t.update_shopid(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_cardid"       { return t.update_cardid(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_category" 	{ return t.update_category(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_cardlevel" 	{ return t.update_cardlevel(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_cardclass" 	{ return t.update_cardclass(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_tel" 			{ return t.update_tel(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_password" 	{ return t.update_password(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_money" 		{ return t.update_money(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_point" 		{ return t.update_point(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_expdate" 		{ return t.update_expdate(stub, card, caller, caller_affiliation, newValue)
		} else if function == "update_expired" 		{ return t.update_expired(stub, card, caller, caller_affiliation, newValue)}

	} else if strings.Contains(function, "transfer_card") == true{
		
		cardid := args[1]
		card, err := t.retrieve_card(stub, cardid)
		if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
						return nil, errors.New("Error retrieving v5c") }

		rec_user := args[2]
		rec_affiliation , err := t.check_affiliation(stub, rec_user)
			if err != nil  { return nil, errors.New("Invalid rec_affiliation value passed") }
												
		vbytes, err := json.Marshal(card)
			if err != nil { return nil, errors.New("ERROR: Invalid card object,cannot Marshal to bytes") }
		fmt.Printf("Card Content: %s", string(vbytes)); 


		if  function == "transfer_card_shop_to_consumer"   { return t.transfer_card_shop_to_consumer(stub, card, caller, caller_affiliation, rec_user, rec_affiliation)
		} else if  function == "transfer_card_consumer_to_consumer" 	   { return t.transfer_card_consumer_to_consumer(stub, card, caller, caller_affiliation, rec_user, rec_affiliation)
		} else if  function == "transfer_card_consumer_to_shop"  { return t.transfer_card_consumer_to_shop(stub, card, caller, caller_affiliation, rec_user, rec_affiliation) }
	
	} else if strings.Contains(function, "_mp_") == true{

		money, err := strconv.Atoi(args[1])
		if err != nil { fmt.Printf("strconv.Atoi args4 money error: ", err); 
						return nil, errors.New("strconv.Atoi args4 money error") }

		point, err := strconv.Atoi(args[2])
		if err != nil { fmt.Printf("strconv.Atoi args4 point error: ", err); 
						return nil, errors.New("strconv.Atoi args4 point error") }

		fmt.Printf("transfer mp start ")
		if  function == "transfer_mp_consumer_to_consumer"   {    //(caller, money, point, sccardid, receiver, tcardid)
				
			sccardid := args[3]
			scard, err := t.retrieve_card(stub, sccardid)
			if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
							return nil, errors.New("Error retrieving v5c") }

			receiver := args[4]		
			tcardid := args[5]
			tcard, err := t.retrieve_card(stub, tcardid)
			if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
							return nil, errors.New("Error retrieving v5c") }
			
			return t.transfer_mp_consumer_to_consumer(stub, money , point , caller , scard, receiver , tcard )

		} else if  function == "deposit_mp_shop_to_consumer" 	   { //(caller, money, point, receiver, tcardid)

			fmt.Printf("deposit_mp_shop_to_consumer 1")
			receiver := args[3]		
			tcardid := args[4]
			tcard, err := t.retrieve_card(stub, tcardid)
			fmt.Printf("deposit_mp_shop_to_consumer 2")
			if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
							return nil, errors.New("Error retrieving v5c") }

			fmt.Printf("get into deposit_mp_shop_to_consumer 3");
			return t.deposit_mp_shop_to_consumer(stub, money , point , caller, receiver , tcard)

		} else if  function == "spend_mp_consumer_to_shop"  {  //(caller, money, point, sccardid, shopid)
			
			fmt.Printf("spend_mp_consumer_to_shop 1")
			sccardid := args[3]
			scard, err := t.retrieve_card(stub, sccardid)
			fmt.Printf("spend_mp_consumer_to_shop 2")
			if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); 
							return nil, errors.New("Error retrieving v5c") }
fmt.Printf("spend_mp_consumer_to_shop 13")

			shopid := args[4]
			fmt.Printf("get into spend_mp_consumer_to_shop");
			return t.spend_mp_consumer_to_shop(stub, money , point , caller , scard, shopid) }
	
	} 

	return nil, errors.New("Function of that name doesn't exist.")

}

//=================================================================================================================================	
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================	
func (t *CardTransactionChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
													
	//caller, caller_affiliation, err := t.get_caller_data(stub)
	//if err != nil { fmt.Printf("QUERY: Error retrieving caller details", err); return nil, errors.New("QUERY: Error retrieving caller details: "+err.Error()) }
	caller := args[0]
 	caller_affiliation , err := t.check_affiliation(stub, caller)
	if err != nil  { return nil, errors.New("Invalid caller_affiliation value") }
															
	if function == "get_users" { 
		caller := args[0]
		return t.get_users(stub, caller)
	} else if function == "get_user_detail" { 
		caller := args[0]
		userName := args[1]
		user,err :=  t.get_user_detail(stub, caller, userName)
		if err != nil { fmt.Printf("get_user_detail error: ", err); 
						return nil, errors.New("get_user_detail error") }

		vbytes, err := json.Marshal(user)
		if err != nil { fmt.Printf("json.Marshal(user): ", err); 
						return nil, errors.New("json.Marshal(user)") }

		return vbytes, nil
	} else if function == "get_card_details" { 
	
			if len(args) != 2 { fmt.Printf("Incorrect number of arguments passed"); 
				return nil, errors.New("QUERY: Incorrect number of arguments passed") }
	
			v, err := t.retrieve_card(stub, args[1])
			if err != nil { fmt.Printf("QUERY: Error retrieving v5c: %s", err); 
				return nil, errors.New("QUERY: Error retrieving v5c "+err.Error()) }
	
			return t.get_card_details(stub, v, caller, caller_affiliation)
			
	} else if function == "get_cards" {
			return t.get_cards(stub, caller, caller_affiliation)

	} else if function == "get_card_templates" {
			return t.get_card_templates(stub, caller, caller_affiliation)
	} else if function == "get_shopLedger" {
			shopid := args[0]
			templateid := args[1]
			return t.get_shopLedger(stub, shopid, templateid)
	}   

	return nil, errors.New("Received unknown function invocation")
}


//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 authority_to_manufacturer
//=================================================================================================================================
/*
func (t *CardTransactionChaincode) kakacenter_to_shop(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	
	if     	v.status				== STATE_TEMPLATE	&&
			v.owner					== caller			&&
			caller_affiliation		== KAKACENTER		&&
			recipient_affiliation	== SHOP				&&
			v.scrapped				== false			{		// If the roles and users are ok 
	
					v.owner  = recipient_name		// then make the owner the new owner
					v.status = STATE_SHOP			// and mark it in the state of manufacture
	
	} else {									// Otherwise if there is an error
	
															fmt.Printf("KAKACENTER_TO_SHOP: Permission Denied");
															return nil, errors.New("Permission Denied")
	
	}
	
	_, err := t.save_card(stub, v)						// Write new state

															if err != nil {	fmt.Printf("KAKACENTER_TO_SHOP: Error saving changes: %s", err); return nil, errors.New("Error saving changes")	}
														
	return nil, nil									// We are Done
	
}
*/


//=================================================================================================================================
//	 card base functions
//=================================================================================================================================
//   save_card - Writes to the ledger the Card struct passed in a JSON format. Uses the shim file's 
//				  method 'PutState'.
//==============================================================================================================================
func (t *CardTransactionChaincode) save_template(stub shim.ChaincodeStubInterface, v Card) (bool, error) {
	 
	bytes, err := json.Marshal(v)
	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting card card record: %s", err); return false, errors.New("Error converting card record") }

	err = stub.PutState(v.Kakaid, bytes)
	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing  card record: %s", err); return false, errors.New("Error storing card record") }

	return true, nil
}

//==============================================================================================================================
// save_card - Writes to the ledger the Card struct passed in a JSON format. Uses the shim file's 
//				  method 'PutState'.
//==============================================================================================================================
func (t *CardTransactionChaincode) save_card(stub shim.ChaincodeStubInterface, v Card) (bool, error) {
	 
	bytes, err := json.Marshal(v)
	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting card card record: %s", err); return false, errors.New("Error converting card record") }

	if( t.checkCardId(v.Cardid) == false ){
		err = stub.PutState(v.Kakaid, bytes)
	}else{
		err = stub.PutState(v.Cardid, bytes)
	}
	
	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing  card record: %s", err); return false, errors.New("Error storing card record") }

	return true, nil
}

//==============================================================================================================================
//	 retrieve_card - Gets the state of the data at v5cID in the ledger then converts it from the stored 
//					JSON into the Card struct for use in the contract. Returns the Card struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *CardTransactionChaincode) retrieve_card(stub shim.ChaincodeStubInterface, cardID string) (Card, error) {
	
	var v Card

	bytes, err := stub.GetState(cardID);					
	if err != nil {	fmt.Printf("RETRIEVE_CARD: Failed to invoke Card_code_cardID: %s", err); return v, errors.New("RETRIEVE_CARD: Error retrieving card with v5cID = " + cardID) }

	err = json.Unmarshal(bytes, &v);						
	if err != nil {	fmt.Printf("RETRIEVE_CARD: Corrupt card record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_CARD: Corrupt card record"+string(bytes))	}
	
	return v, nil
}

//==============================================================================================================================
//	 retrieve_card - Gets the state of the data at v5cID in the ledger then converts it from the stored 
//					JSON into the Card struct for use in the contract. Returns the Card struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *CardTransactionChaincode) get_Shopid(stub shim.ChaincodeStubInterface, caller string) (string) {

	return caller
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================									
//	 Create Card Template- 								
//=================================================================================================================================

func (t *CardTransactionChaincode) create_card_template(stub shim.ChaincodeStubInterface, caller string, caller_affiliation int, templateID string) ([]byte, error) {								

	fmt.Printf("start  create_card_template \n ");

	// build Card obejct by json
	var v Card																																									

	kakaid          := "\"Kakaid\":\""+templateID+"\", "	
	cardid          := "\"Cardid\":\"UNDEFINED\", "							
	shop            := "\"Shop\":\"UNDEFINED\", "
	shopid          := "\"Shopid\":\"UNDEFINED\", "
	category        := "\"Category\":\"UNDEFINED\", "
	cardlevel       := "\"Cardlevel\":\"UNDEFINED\", "
	cardclass       := "\"Cardclass\":\"UNDEFINED\", "
	owner           := "\"Owner\":\""+caller+"\", "
	tel             := "\"Tel\":\"UNDEFINED\", "
	password   	    := "\"Password\":\"UNDEFINED\", "
	money           := "\"Money\":0, "
	point           := "\"Point\":0, "
	releasedate     := "\"Releasedate\":\"UNDEFINED\", "
	expdate         := "\"Expdate\":\"UNDEFINED\", "
	getdate         := "\"Getdate\":\"UNDEFINED\", "
	expired			:= "\"Expired\":false, "
	scrapped       	:= "\"Scrapped\":false, "
	status       	:= "\"Status\":0 "

	card_json := "{"+kakaid+cardid+shop+shopid+category+cardlevel+cardclass+owner+tel+password+money+point+releasedate+expdate+getdate+expired+scrapped+status+"}" 	// Concatenates the variables to create the total JSON object
	
	fmt.Printf("test json: %s ",card_json);

	err := json.Unmarshal([]byte(card_json), &v)		//  new card json -> Card Object
	fmt.Printf("test 04 ");
		if err != nil { 
			fmt.Printf("test err is not nil , err is : %s",err);
			return nil, errors.New("Invalid JSON object") 
			
			}

	fmt.Printf("json to card tempalte object ");
	fmt.Printf("json to card tempalte object :  kakaid = %s", v.Kakaid);
	fmt.Printf("json to card tempalte object :  cardid = %s", v.Cardid);
	fmt.Printf("json to card tempalte object :  owner = %s", v.Owner);

	//if auth to create template
	if 	caller_affiliation != KAKACENTER {							// Only the regulator can create a new v5c
		return nil, errors.New("Permission Denied")
	}

	matched, err := regexp.Match("^[A-z][A-z][A-z]", []byte(templateID))  	// 2 char + 5 digits
		if err != nil  || matched ==false { fmt.Printf("CREATE_CARD: Invalid cardID: %s", err); return nil, errors.New("Invalid v5cID") }
	
	fmt.Printf("test 0 : %s",templateID);
	fmt.Printf("test 0 ");
	record, err := stub.GetState(templateID) 			// check if card already exists
	fmt.Printf("test 1 ");
		if record != nil { return nil, errors.New("Card already exists") }
	fmt.Printf("test 2 ");

	//save template
	_, err  = t.save_template(stub, v)									
		if err != nil { fmt.Printf("CREATE_CARD_TEMPLATE: Error saving changes: %s", err); 
						return nil, errors.New("Error saving changes") }
	
	fmt.Printf("save tamplate ok");

	//save cardID in CARD_TEMPLATE_HOLDER
	bytes, err := stub.GetState(CARD_TEMPLATE_HOLDER)
		if err != nil { return nil, errors.New("Unable to get cardKakaIDs") }
	
	fmt.Print("get state this template  : %s", string(bytes))
																	
	var card_template_holder Card_Holder
	
	if len(bytes)!=0 {
		fmt.Print(" template holder have data : %s", string(bytes))
		
		err = json.Unmarshal(bytes, &card_template_holder)
		if err != nil {	return nil, errors.New("Corrupt Card_Template_Holder record") }
	}
	

	card_template_holder.Cards = append(card_template_holder.Cards, templateID)
	
	fmt.Print("Marshal, holder num: %s ", string(len(card_template_holder.Cards)))
	bytes, err = json.Marshal(card_template_holder)
		if err != nil { fmt.Print("Error creating Card_Template_Holder record") }

	fmt.Print("put state this template : %s", string(bytes))

	err = stub.PutState(CARD_TEMPLATE_HOLDER, bytes)
		if err != nil { return nil, errors.New("Unable to put the CARD_TEMPLATE_HOLDER state") }
	
	return nil, nil

}


//=================================================================================================================================
//	 transfer temaplte from kakacenter to shop
//=================================================================================================================================
func (t *CardTransactionChaincode) transfer_template_to_shop(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {


	if      v.Kakaid	== "UNDEFINED" {					//If key part of the card is undefined it has not been fully manufacturered so cannot be sent
					fmt.Printf("SHOP_TO_CONSUMER: Car not fully defined")
					return nil, errors.New("Car not fully defined")
	}
	
	if 		v.Status				== STATE_TEMPLATE	&& 
			v.Owner					== caller		&& 
			caller_affiliation		== KAKACENTER	&&
			recipient_affiliation	== SHOP			&& 
			v.Scrapped     == false							{
			
					v.Owner = recipient_name
					v.Status = STATE_SHOP

					timestamp := time.Now().Unix()
					tm := time.Unix(timestamp, 0)
					v.Releasedate = tm.Format("2006-01-02 03:04:05 PM")  //  str to timestmap: tm2, _ := time.Parse("01/02/2006", releasedate)
					v.Getdate = v.Releasedate
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_template(stub, v)
	
															if err != nil { fmt.Printf("SHOP_TO_CONSUMER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}
/*
func (t *CardTransactionChaincode) update_template(stub shim.ChaincodeStubInterface, caller string, caller_affiliation int,cardTemplateId string, updatedCardJson string)([]byte, error) {
	
	if 		v.Status				== STATE_TEMPLATE	&& 
			v.Owner					== caller		&& 
			(caller_affiliation		== KAKACENTER || caller_affiliation		== SHOP	) 	&&
			v.Scrapped     == false							{
			
			
			cardTemplate, err := t.retrieve_card(stub, cardTemplateId)
		
				if err != nil {return nil, errors.New("Failed to retrieve V5C")}

			err = json.Unmarshal([]byte(updatedCardJson), &cardTemplate)		//  new card json -> Card Object
				if err != nil { return nil, errors.New("Invalid JSON object") }

			_, err  = t.save_template(stub, cardTemplate)									
				if err != nil { fmt.Printf("CREATE_CARD_TEMPLATE: Error saving changes: %s", err); 
						return nil, errors.New("Error saving changes") }
	
	} else {
															return nil, errors.New("Permission denied")
	}
	return nil,nil
}
*/
//=================================================================================================================================									
//	 Create Card Template- 								
//=================================================================================================================================
func (t *CardTransactionChaincode) generate_card_id(cardTemplate_KakaIDs string, pos int) (string) {
	basenum := 1000000

	return cardTemplate_KakaIDs + "-" +"A" + strconv.Itoa(basenum + pos)
}

func (t *CardTransactionChaincode) checkCardId(cardId string) (bool) {
	return strings.Contains(cardId, "-")
}

func (t *CardTransactionChaincode) create_card_by_template(stub shim.ChaincodeStubInterface, caller string, caller_affiliation int, cardTemplate_KakaIDs string, cardNum int) ([]byte, error) {								

	if 	caller_affiliation != SHOP {							// Only the regulator can create a new v5c
		return nil, errors.New("Permission Denied")
	}

	matched, err := regexp.Match("^[A-z][A-z][A-z]", []byte(cardTemplate_KakaIDs))  	// 2 char + 5 digits
		if err != nil   || matched ==false { fmt.Printf("CREATE_CARD: Invalid cardID: %s", err); return nil, errors.New("Invalid v5cID") }
	
	cardTemplateBytes, err := stub.GetState(cardTemplate_KakaIDs) 			// check if card template exists
		if cardTemplateBytes == nil { return nil, errors.New("Card temaplte is not exists") }
	
	var cardTemplate Card	
		err = json.Unmarshal(cardTemplateBytes, &cardTemplate)		//  new card json -> Card Object
			if err != nil { return nil, errors.New("------------Invalid cardTemplateBytes JSON object") }

	
	if 		cardTemplate.Shopid 	 	== "UNDEFINED" || 					
			cardTemplate.Shop  			== "UNDEFINED" || 
			cardTemplate.Cardclass 		== "UNDEFINED" || 
			cardTemplate.Expdate		 == "UNDEFINED"  {					//If key part of the card is undefined it has not been fully manufacturered so cannot be sent
															fmt.Printf("SHOP_TO_CONSUMER: Car template not fully defined")
															return nil, errors.New("Car template not fully defined")
	}
	
	// create new Card by template
	//var card Card	
	//err = json.Unmarshal(cardTemplateBytes, &card)		//  new card json -> Card Object
	//	if err != nil { return nil, errors.New("Invalid JSON object") }


	//once create new card, create or update shop ledger, 
	var	shopLedger ShopLedger
	shopid := t.get_Shopid(stub, caller)
	shopLedgerBytes ,err := t.get_shopLedger(stub, shopid, cardTemplate_KakaIDs)
	if shopLedgerBytes == nil {
		shopLedgerBytes ,err = t.add_new_shopLedger(stub, shopid, cardTemplate_KakaIDs)
	}

	err = json.Unmarshal([]byte(shopLedgerBytes), &shopLedger)	
	if err != nil { return nil, errors.New("------------Invalid shopLedgerBytes JSON object") }


	// get Card_Holder
	var card_holder Card_Holder
	bytes, err := stub.GetState(CARD_HOLDER)
		if err != nil { return nil, errors.New("Unable to get cardKakaIDs") }	

	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &card_holder)
		if err != nil {	return nil, errors.New("Corrupt Card_Template_Holder record") }
	}															
	

	for cardindex := shopLedger.CardIdIndex;  cardindex < shopLedger.CardIdIndex + cardNum ;cardindex++ {
		//create new card from template
		var card Card	
		err = json.Unmarshal(cardTemplateBytes, &card)		//  new card json -> Card Object
			if err != nil { return nil, errors.New("Invalid JSON object") }

		//err = json.Unmarshal(cardTemplateBytes, &card)		//  new card json -> Card Object
		//	if err != nil { return nil, errors.New("Invalid JSON object") }
		card.Cardid = t.generate_card_id(cardTemplate_KakaIDs, cardindex + 1 )

		fmt.Printf("CREATE_CARD cardid:  %s", card.Cardid);
		card.Kakaid 		 = 	cardTemplate_KakaIDs	
		fmt.Printf("CREATE_CARD Kakaid: %s", card.Kakaid);

		//save card to state
		_, err  = t.save_card(stub, card)									
		if err != nil { fmt.Printf("CREATE_CARD_TEMPLATE: Error saving changes: %s", err); 
						return nil, errors.New("Error saving changes") }
		
		//add to card_holder
		card_holder.Cards = append(card_holder.Cards, card.Cardid)
		fmt.Printf("Append 1 cardid  \n");
	}
	fmt.Printf("Marshal card holder to bytes");
	//save cardIDs in CARD_HOLDER
	bytes, err = json.Marshal(card_holder)
		if err != nil { fmt.Print("Error creating Card_Holder record") }

	fmt.Printf("PutState card to card_holder");
	err = stub.PutState(CARD_HOLDER, bytes)
		if err != nil { return nil, errors.New("Unable to put the CARD_HOLDER state") }
	fmt.Printf("PutState ok");


	//update shop ledger
	shopLedger.Shopid = shopid 
	shopLedger.Qty = shopLedger.Qty + cardNum
	shopLedger.CardIdIndex = shopLedger.CardIdIndex + cardNum
	shopLedger.InitMoney = shopLedger.InitMoney + cardTemplate.Money * cardNum
	shopLedger.InitPoint = shopLedger.InitPoint + cardTemplate.Point * cardNum
	t.update_shopLedger(stub, caller, cardTemplate_KakaIDs, shopLedger)

	fmt.Printf("Put ShopLedger ok");

	return nil, nil

}

//=================================================================================================================================
//	 manufacturer_to_private
//=================================================================================================================================
func (t *CardTransactionChaincode) request_card_by_template(stub shim.ChaincodeStubInterface, caller string, caller_affiliation int, cardTemplate_KakaIDs string) ([]byte, error) {								

	if 	caller_affiliation !=  CONSUMER {							// Only the regulator can create a new v5c
		return nil, errors.New("Permission Denied")
	}

	matched, err := regexp.Match("^[A-z][A-z][A-z]", []byte(cardTemplate_KakaIDs))  	// 2 char + 5 digits
		if err != nil   || matched ==false { fmt.Printf("CREATE_CARD: Invalid cardID: %s", err); return nil, errors.New("Invalid v5cID") }
	
	cardTemplateBytes, err := stub.GetState(cardTemplate_KakaIDs) 			// check if card template exists
		if cardTemplateBytes == nil { return nil, errors.New("Card temaplte is not exists") }
	


	//once create new card, create or update shop ledger, 
	var	shopLedger ShopLedger
	shopid := t.get_Shopid(stub, caller)
	shopLedgerBytes ,err := t.get_shopLedger(stub, shopid, cardTemplate_KakaIDs)
	if shopLedgerBytes == nil {
		shopLedgerBytes ,err = t.add_new_shopLedger(stub, shopid, cardTemplate_KakaIDs)
	}

	err = json.Unmarshal([]byte(shopLedgerBytes), &shopLedger)	
	if err != nil { return nil, errors.New("------------Invalid shopLedgerBytes JSON object") }


	// get Card_Holder
	var card_holder Card_Holder
	bytes, err := stub.GetState(CARD_HOLDER)
		if err != nil { return nil, errors.New("Unable to get cardKakaIDs") }	

	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &card_holder)
		if err != nil {	return nil, errors.New("Corrupt Card_Template_Holder record") }
	}															
	

	//create new card from template
	var card Card	
	err = json.Unmarshal(cardTemplateBytes, &card)		//  new card json -> Card Object
		if err != nil { return nil, errors.New("Invalid JSON object") }

	//err = json.Unmarshal(cardTemplateBytes, &card)		//  new card json -> Card Object
	//	if err != nil { return nil, errors.New("Invalid JSON object") }
	card.Cardid = t.generate_card_id(cardTemplate_KakaIDs, shopLedger.CardIdIndex + 1 )

	fmt.Printf("CREATE_CARD cardid:  %s", card.Cardid);
	card.Kakaid 		 = 	cardTemplate_KakaIDs	
	fmt.Printf("CREATE_CARD Kakaid: %s", card.Kakaid);

	card.Owner = caller
	card.Status = STATE_CONSUMER_OWNERSHIP

	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	card.Releasedate = tm.Format("2006-01-02 03:04:05 PM")  //  str to timestmap: tm2, _ := time.Parse("01/02/2006", releasedate)
	card.Getdate = card.Releasedate

	
	//save card to state
	_, err  = t.save_card(stub, card)									
	if err != nil { fmt.Printf("CREATE_CARD_TEMPLATE: Error saving changes: %s", err); 
					return nil, errors.New("Error saving changes") }
	
	//add to card_holder
	card_holder.Cards = append(card_holder.Cards, card.Cardid)

	fmt.Printf("Marshal card holder to bytes");
	//save cardIDs in CARD_HOLDER
	bytes, err = json.Marshal(card_holder)
		if err != nil { fmt.Print("Error creating Card_Holder record") }

	fmt.Printf("PutState card to card_holder");
	err = stub.PutState(CARD_HOLDER, bytes)
		if err != nil { return nil, errors.New("Unable to put the CARD_HOLDER state") }
	fmt.Printf("PutState ok");


	//update shop ledger
	shopLedger.Shopid = shopid
	shopLedger.Qty = shopLedger.Qty + 1
	shopLedger.CardIdIndex = shopLedger.CardIdIndex + 1
	shopLedger.InitMoney = shopLedger.InitMoney + card.Money
	shopLedger.InitPoint = shopLedger.InitPoint + card.Point
	t.update_shopLedger(stub, shopid, cardTemplate_KakaIDs, shopLedger)

	fmt.Printf("Put ShopLedger ok");

	return nil, nil

}

//=================================================================================================================================
//	 manufacturer_to_private
//=================================================================================================================================
func (t *CardTransactionChaincode) transfer_card_shop_to_consumer(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {


	if 		v.Shop 	 	== "UNDEFINED" || 					
			v.Cardid  	== "UNDEFINED" || 
			v.Cardlevel == "UNDEFINED" || 
			v.Cardclass == "UNDEFINED" || 
			v.Expdate == "UNDEFINED"  {					//If key part of the card is undefined it has not been fully manufacturered so cannot be sent
															fmt.Printf("SHOP_TO_CONSUMER: Car not fully defined")
															return nil, errors.New("Car not fully defined")
	}
	
	if 		v.Status				== STATE_SHOP	&& 
			v.Owner					== caller		&& 
			caller_affiliation		== SHOP			&&
			recipient_affiliation	== CONSUMER		&& 
			v.Scrapped     == false							{
			
					v.Owner = recipient_name
					v.Status = STATE_CONSUMER_OWNERSHIP

					timestamp := time.Now().Unix()
					tm := time.Unix(timestamp, 0)
					v.Releasedate = tm.Format("2006-01-02 03:04:05 PM")  //  str to timestmap: tm2, _ := time.Parse("01/02/2006", releasedate)
					v.Getdate = v.Releasedate
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("SHOP_TO_CONSUMER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 private_to_private
//=================================================================================================================================
func (t *CardTransactionChaincode) transfer_card_consumer_to_consumer(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
	

	if 		v.Status				== STATE_CONSUMER_OWNERSHIP	&&
			v.Owner					== caller					&&
			caller_affiliation		== CONSUMER			&& 
			recipient_affiliation	== CONSUMER			&&
			v.Scrapped				== false					{
			
					v.Owner = recipient_name
					timestamp := time.Now().Unix()
					tm := time.Unix(timestamp, 0)
					v.Getdate = tm.Format("2006-01-02 03:04:05 PM")  //  str to timestmap: tm2, _ := time.Parse("01/02/2006", releasedate)
					
					
	} else {
		
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("CONSUMER_TO_CONSUMER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 private_to_lease_company
//=================================================================================================================================
func (t *CardTransactionChaincode) transfer_card_consumer_to_shop(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {

	if 		v.Status				== STATE_CONSUMER_OWNERSHIP	&& 
			v.Owner					== caller					&& 
			caller_affiliation		== CONSUMER					&& 
			recipient_affiliation	== SHOP						&& 
			v.Scrapped     			== false					{
		
					v.Owner = recipient_name
					timestamp := time.Now().Unix()
					tm := time.Unix(timestamp, 0)
					v.Getdate = tm.Format("2006-01-02 03:04:05 PM")  //  str to timestmap: tm2, _ := time.Parse("01/02/2006", releasedate)
					
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
															if err != nil { fmt.Printf("consumer_identityCard_to_shop: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}


func (t *CardTransactionChaincode) transfer_mp_shop_to_consumer(stub shim.ChaincodeStubInterface, money int, point int, caller string, sc Card, receiver string, tc Card) ([]byte, error) {

		return nil, errors.New("not implemented")
}

func (t *CardTransactionChaincode) transfer_mp_consumer_to_shop(stub shim.ChaincodeStubInterface, money int, point int, caller string, sc Card, receiver string, tc Card) ([]byte, error) {

		return nil, errors.New("not implemented")
}
//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 transfer_mp_consumer_to_consumer
//=================================================================================================================================
func (t *CardTransactionChaincode) transfer_mp_consumer_to_consumer(stub shim.ChaincodeStubInterface, money int, point int, caller string, sc Card, receiver string, tc Card) ([]byte, error) {
fmt.Printf("start transfer_mp_consumer_to_consumer")
	if sc.Money < money || sc.Point < point{
		fmt.Printf("money or point is not enough")
		return nil, errors.New("card asset is not enough")
	}
fmt.Printf("test 1")
	caller_affiliation , _ := t.check_affiliation(stub, caller)
	fmt.Printf("test 2")
	receiver_affiliation , _ := t.check_affiliation(stub, receiver)
	fmt.Printf("test 3")

	if		sc.Status				== STATE_CONSUMER_OWNERSHIP	&&
			sc.Owner  				== caller					&& 
			sc.Scrapped  			== false					&& 
			sc.Expired  			== false					&& 

			tc.Status				== STATE_CONSUMER_OWNERSHIP	&&
			tc.Owner  				== receiver					&& 
			sc.Scrapped  			== false					&& 
			sc.Expired  			== false					&&

			sc.Shopid 				== tc.Shopid 					&&

			caller_affiliation		== CONSUMER			&& 
			receiver_affiliation	== CONSUMER			{
		
				fmt.Printf("add and substract")
				tc.Money = tc.Money + money
				sc.Money = sc.Money - money
				tc.Point = tc.Point + point
				sc.Point = sc.Point - point
				//add event to triger the shop db update
	
	} else {
			fmt.Printf("Permission denied----------------------------")
			return nil, errors.New("Permission denied")
	}
	
	fmt.Printf("---------------save_card sc---------------------------")
    _, err1 := t.save_card(stub, sc)
   fmt.Printf("---------------save_card tc---------------------------")
    _, err2 := t.save_card(stub, tc)
		fmt.Printf("---------------save_card ok---------------------------")		
		if err1 != nil || err2 != nil { fmt.Printf("transactionCard_consumer_to_consumer: Error saving changes: %s", err1); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}


//=================================================================================================================================
//	 deposit_mp_shop_to_consumer
//=================================================================================================================================
func (t *CardTransactionChaincode) deposit_mp_shop_to_consumer(stub shim.ChaincodeStubInterface, money int, point int, caller string, receiver string, tc Card) ([]byte, error) {

	fmt.Printf("start deposit_mp_shop_to_consumer")

	caller_affiliation , _ := t.check_affiliation(stub, caller)
	receiver_affiliation , _ := t.check_affiliation(stub, receiver)
	fmt.Printf("test 3")

	shopid := t.get_Shopid(stub, caller)
	// update shop ledger, 
			var	shopLedger ShopLedger
			shopLedgerBytes ,err := t.get_shopLedger(stub, shopid, tc.Kakaid)
			if shopLedgerBytes == nil {
				fmt.Printf("Permission denied----------------------------")
					return nil, errors.New("Permission denied")
			}

			err = json.Unmarshal([]byte(shopLedgerBytes), &shopLedger)	
			if err != nil { return nil, errors.New("------------Invalid shopLedgerBytes JSON object") }



	if		tc.Status				== STATE_CONSUMER_OWNERSHIP	&&
			tc.Owner  				== receiver					&& 
			tc.Scrapped  			== false					&& 
			tc.Expired  			== false					&&

			tc.Shopid 				== shopLedger.Shopid 			&&

			caller_affiliation		== SHOP			&& 
			receiver_affiliation	== CONSUMER			{
		
				fmt.Printf("add and substract")
				tc.Money = tc.Money + money
				tc.Point = tc.Point + point
	} else {
			fmt.Printf("Permission denied----------------------------")
			return nil, errors.New("Permission denied")
	}

   fmt.Printf("---------------save_card tc---------------------------")
    _, err = t.save_card(stub, tc)

	//save shop ledger
			shopLedger.DepositMoney = shopLedger.DepositMoney + money
			shopLedger.DepositPoint = shopLedger.DepositPoint + point
			t.update_shopLedger(stub, shopid, tc.Kakaid, shopLedger)

			fmt.Printf("Put ShopLedger ok");

		fmt.Printf("---------------save_card ok---------------------------")		
		if err != nil  { fmt.Printf("transactionCard_consumer_to_consumer: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 spend_mp_consumer_to_consumer
//=================================================================================================================================
func (t *CardTransactionChaincode) spend_mp_consumer_to_shop(stub shim.ChaincodeStubInterface, money int, point int, caller string, sc Card, shopid string) ([]byte, error) {
	
	fmt.Printf("start spend_mp_consumer_to_consumer")
	if sc.Money < money || sc.Point < point{
		fmt.Printf("money or point is not enough")
		return nil, errors.New("card asset is not enough")
	}

// update shop ledger, 
			var	shopLedger ShopLedger
			shopLedgerBytes ,err := t.get_shopLedger(stub, shopid, sc.Kakaid)
			if shopLedgerBytes == nil {
				fmt.Printf("pay to wrong shop ,please check shop name--")
					return nil, errors.New("pay to wrong shop ,please check shop name ")
			}

			err = json.Unmarshal([]byte(shopLedgerBytes), &shopLedger)	
			if err != nil { return nil, errors.New("------------Invalid shopLedgerBytes JSON object") }


	fmt.Printf("test 1")
	caller_affiliation , _ := t.check_affiliation(stub, caller)
	fmt.Printf("test 3")

			

	if		sc.Status				== STATE_CONSUMER_OWNERSHIP	&&
			sc.Owner  				== caller					&& 
			sc.Scrapped  			== false					&& 
			sc.Expired  			== false					&& 

			sc.Shopid 				== shopLedger.Shopid 			&&

			caller_affiliation		== CONSUMER		{
		
				fmt.Printf("add and substract")
				sc.Money = sc.Money - money
				sc.Point = sc.Point - point	
	} else {
			fmt.Printf("Permission denied----------------------------")
			return nil, errors.New("Permission denied")
	}
	
	fmt.Printf("---------------save_card sc---------------------------")
    _, err = t.save_card(stub, sc)
  
  			shopLedger.ConsumeMoney = shopLedger.ConsumeMoney + money
			shopLedger.ConsumePoint = shopLedger.ConsumePoint + point
			t.update_shopLedger(stub, shopid, sc.Kakaid, shopLedger)

			fmt.Printf("Put ShopLedger ok");
		fmt.Printf("---------------save_card ok---------------------------")		
		if err != nil { fmt.Printf("transactionCard_consumer_to_consumer: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 Update Functions
//=================================================================================================================================
//	 update_money
//=================================================================================================================================

func (t *CardTransactionChaincode) update_money(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	new_money, err := strconv.Atoi(new_value) 		                // will return an error if the new vin contains non numerical chars
		if err != nil  { return nil, errors.New("Invalid value passed for money") }
	
	if 		//v.status			== STATE_SHOP	&& 
			//v.owner				== caller				&&
			caller_affiliation	== SHOP			&&
			//v.VIN				== 0					&&			// Can't change the VIN after its initial assignment
			v.Scrapped			== false				{
			
					v.Money = new_money					// Update to the new value
	} else {
	
															return nil, errors.New("Permission denied")
		
	}
	
	_, err  = t.save_card(stub, v)						// Save the changes in the blockchain
	
															if err != nil { fmt.Printf("update_money: Error saving changes: %s", err); return nil, errors.New("Error saving changes") } 
	
	return nil, nil
	
}



//=================================================================================================================================
//	 update_point
//=================================================================================================================================

func (t *CardTransactionChaincode) update_point(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	new_point, err := strconv.Atoi(new_value)		                // will return an error if the new vin contains non numerical chars
	if err != nil { return nil, errors.New("Invalid value passed for point") }
	
	if 		//v.status			== STATE_SHOP	&& 
			//v.owner				== caller				&&
			caller_affiliation	== SHOP			&&
			//v.VIN				== 0					&&			// Can't change the VIN after its initial assignment
			v.Scrapped			== false				{
			
					v.Point = new_point					// Update to the new value
	} else {
	
															return nil, errors.New("Permission denied")
		
	}
	
	_, err  = t.save_card(stub, v)						// Save the changes in the blockchain
	
															if err != nil { fmt.Printf("update_point: Error saving changes: %s", err); return nil, errors.New("Error saving changes") } 
	
	return nil, nil
	
}


//	 only can be update by shop
//=================================================================================================================================
//	 update_shop
//=================================================================================================================================

func (t *CardTransactionChaincode) update_shopid(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Status			== STATE_SHOP	&&
			v.Owner				== caller				&& 
			caller_affiliation	== SHOP			&&
			v.Scrapped			== false				{
			
					v.Shopid = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_shopid: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}


//=================================================================================================================================
//	 update_shop
//=================================================================================================================================

func (t *CardTransactionChaincode) update_shop(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Status			== STATE_SHOP	&&
			v.Owner				== caller				&& 
			caller_affiliation	== SHOP			&&
			v.Scrapped			== false				{
			
					v.Shop = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_shop: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_cardid
//=================================================================================================================================

func (t *CardTransactionChaincode) update_cardid(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Status			== STATE_SHOP	&&
			v.Owner				== caller				&& 
			caller_affiliation	== SHOP			&&
			v.Scrapped			== false				{
			
					v.Cardid = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_cardid: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_cardlevel
//=================================================================================================================================


func (t *CardTransactionChaincode) update_cardlevel(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Status			== STATE_SHOP	&&
			v.Owner				== caller				&& 
			caller_affiliation	== SHOP			&&
			v.Scrapped			== false				{
			
					v.Cardlevel = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_cardlevel: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_cardclass
//=================================================================================================================================

func (t *CardTransactionChaincode) update_cardclass(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Status			== STATE_SHOP	&&
			v.Owner				== caller				&& 
			caller_affiliation	== SHOP			&&
			v.Scrapped			== false				{
			
					v.Cardclass = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_cardclass: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_password
//=================================================================================================================================

func (t *CardTransactionChaincode) update_password(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {
	
	if 		v.Owner				== caller				&& 
			v.Scrapped			== false				{
			
					v.Password = new_value
	} else {
	
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_password: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}


//=================================================================================================================================
//	 update_expdate
//=================================================================================================================================
func (t *CardTransactionChaincode) update_expdate(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {

	if		v.Status			== STATE_SHOP	&&
			v.Owner				== caller		&& 
			caller_affiliation	== SHOP			&&
			v.Scrapped			== false				{
		
					v.Expdate = new_value
				
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_expdate: Error saving changes: %s", err); return nil, errors.New("SCRAP_CARD Error saving changes") }
	
	return nil, nil
	
}


// all role can do this
//=================================================================================================================================
//	 scrap_card
//=================================================================================================================================
func (t *CardTransactionChaincode) scrap_card(stub shim.ChaincodeStubInterface, v Card, caller string) ([]byte, error) {

	if		v.Owner				== caller				&& 
			v.Scrapped			== false				{
		
					v.Scrapped = true
				
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("SCRAP_CARD: Error saving changes: %s", err); return nil, errors.New("SCRAP_CARD Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 update_expired
//=================================================================================================================================
func (t *CardTransactionChaincode) update_expired(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, isexpired string) ([]byte, error) {

	if		v.Owner				== caller				&& 
			v.Scrapped			== false {
				if(isexpired =="true"){
					v.Expired = true
				}else if(isexpired =="false"){
					v.Expired = false
				}else { fmt.Printf("update_expired: value is not true or false"); 
						return nil, errors.New("update_expired: value is not true or false") }
	
					
				
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
		if err != nil { fmt.Printf("update_expired: Error saving changes: %s", err); return nil, errors.New("SCRAP_CARD Error saving changes") }
	
	return nil, nil
	
}


//=================================================================================================================================
//	 update_category
//=================================================================================================================================
func (t *CardTransactionChaincode) update_category(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {

	if		v.Status			== STATE_CONSUMER_OWNERSHIP	&&
			v.Owner				== caller			&& 
			caller_affiliation	== CONSUMER			&&
			v.Scrapped			== false				{
		
					v.Category = new_value
				
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_category: Error saving changes: %s", err); return nil, errors.New("SCRAP_CARD Error saving changes") }
	
	return nil, nil
	
}


//=================================================================================================================================
//	 update_tel
//=================================================================================================================================
func (t *CardTransactionChaincode) update_tel(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int, new_value string) ([]byte, error) {

	if		v.Status			== STATE_CONSUMER_OWNERSHIP	&&
			v.Owner				== caller		&& 
			caller_affiliation	== CONSUMER			&&
			v.Scrapped			== false				{
		
					v.Tel = new_value
				
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_card(stub, v)
	
															if err != nil { fmt.Printf("update_tel: Error saving changes: %s", err); return nil, errors.New("SCRAP_CARD Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
// these 4 item can not been update seperatly. they should be changed in transaction func
//  update_owner
//  update_releasedate
//  update_getdate
//  update_status
//=================================================================================================================================


//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_card_details
//=================================================================================================================================
func (t *CardTransactionChaincode) get_card_details(stub shim.ChaincodeStubInterface, v Card, caller string, caller_affiliation int) ([]byte, error) {
	
	bytes, err := json.Marshal(v)
	
																if err != nil { return nil, errors.New("GET_CARD_DETAILS: Invalid card object") }
																
	if 		v.Owner				== caller		||
			caller_affiliation	== KAKACENTER	||
			caller_affiliation	== SHOP 	{
			
					return bytes, nil		
	} else {
																return nil, errors.New("Permission Denied")	
	}

}


//=================================================================================================================================
//	 get_card_templates
//=================================================================================================================================

func (t *CardTransactionChaincode) get_card_templates(stub shim.ChaincodeStubInterface, caller string, caller_affiliation int) ([]byte, error) {

	bytes, err := stub.GetState(CARD_TEMPLATE_HOLDER)
		
																			if err != nil { return nil, errors.New("Unable to get v5cIDs") }
																	
	var card_template_holder Card_Holder
	
	err = json.Unmarshal(bytes, &card_template_holder)						
	
																			if err != nil {	return nil, errors.New("Corrupt V5C_Holder") }
	
	result := "["
	
	var temp []byte
	var v Card
	
	for _, cardId := range card_template_holder.Cards {
		
		v, err = t.retrieve_card(stub, cardId)
		
		if err != nil {return nil, errors.New("Failed to retrieve V5C")}
		
		temp, err = t.get_card_details(stub, v, caller, caller_affiliation)
		
		if err == nil {
			result += string(temp) + ","	
		}
	}
	
	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}
	
	return []byte(result), nil
}

//=================================================================================================================================
//	 get_cards
//=================================================================================================================================

func (t *CardTransactionChaincode) get_cards(stub shim.ChaincodeStubInterface, caller string, caller_affiliation int) ([]byte, error) {

	bytes, err := stub.GetState(CARD_HOLDER)
		
																			if err != nil { return nil, errors.New("Unable to get v5cIDs") }
																	
	var cardKakaIDs Card_Holder
	
	err = json.Unmarshal(bytes, &cardKakaIDs)						
	
																			if err != nil {	return nil, errors.New("Corrupt V5C_Holder") }
	
	result := "["
	
	var temp []byte
	var v Card
	
	for _, card := range cardKakaIDs.Cards {
		
		v, err = t.retrieve_card(stub, card)
		
		if err != nil {return nil, errors.New("Failed to retrieve V5C")}
		
		temp, err = t.get_card_details(stub, v, caller, caller_affiliation)
		
		if err == nil {
			result += string(temp) + ","	
		}
	}
	
	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}
	
	return []byte(result), nil
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(CardTransactionChaincode))
	
															if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}
