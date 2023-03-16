package messages

type Algo int

const (
	EDDSAKEYGEN1      = "EDDSAKGRound1Message"
	EDDSAKEYGEN2a     = "EDDSAKGRound2Message1"
	EDDSAKEYGEN2b      = "EDDSAKGRound2Message2"
	EDDSAKEYSIGN1      = "EDDSASignRound1Message"
	EDDSAKEYSIGN2      = "EDDSASignRound2Message"
	EDDSAKEYSIGN3      = "EDDSASignRound3Message"
	EDDSAKEYREGROUP1   = "EDDSADGRound1Message"
	EDDSAKEYREGROUP2   = "EDDSADGRound2Message"
	EDDSAKEYREGROUP3a  = "EDDSADGRound3Message1"
	EDDSAKEYREGROUP3b  = "EDDSADGRound3Message2"
	EDDSAKEYREGROUP4   = "EDDSADGRound4Message"
	EDDSAKEYGENROUNDS  = 3
	EDDSAKEYSIGNROUNDS = 3
	EDDSAREGROUPROUNDS = 5

	ECDSAKEYGENROUNDS  = 4
	ECDSAKEYSIGNROUNDS = 8
	ECDSAREGROUPROUNDS = 5

	KEYGEN1            = "KGRound1Message"
	KEYGEN2aUnicast    = "KGRound2Message1"
	KEYGEN2b           = "KGRound2Message2"
	KEYGEN3            = "KGRound3Message"
	KEYSIGN1aUnicast   = "SignRound1Message1"
	KEYSIGN1b          = "SignRound1Message2"
	KEYSIGN2Unicast    = "SignRound2Message"
	KEYSIGN3           = "SignRound3Message"
	KEYSIGN4           = "SignRound4Message"
	KEYSIGN5           = "SignRound5Message"
	KEYSIGN6           = "SignRound6Message"
	KEYSIGN7           = "SignRound7Message"
	KEYSIGN8           = "SignRound8Message"
	KEYSIGN9           = "SignRound9Message"

	KEYREGROUP1      = "DGRound1Message"
	KEYREGROUP2a     = "DGRound2Message1"
	KEYREGROUP2b     = "DGRound2Message2"
	KEYREGROUP3a     = "DGRound3Message1"
	KEYREGROUP3b     = "DGRound3Message2"
	KEYREGROUP4      = "DGRound4Message"

	ECDSAKEYGEN Algo = iota
	ECDSAKEYSIGN
	EDDSAKEYGEN
	EDDSAKEYSIGN
	EDDSAKEYREGROUP
	ECDSAKEYREGROUP
)
