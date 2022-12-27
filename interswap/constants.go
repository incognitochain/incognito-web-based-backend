package interswap

const IncNetworkStr = "inc"
const PAppStr = "papp"
const InterSwapStr = "interswap"

const pDEXType = 1
const pAppType = 2

// path type
const PdexToPApp = 1
const PAppToPdex = 2

const DefaultDecimal = 9

var InterSwapStatus = map[string]int{}

//TODO
const (
	SubmitFailed    int = 0
	FirstPending        = 1
	FirstRefunding      = 2
	FirstRefunded       = 3
	MidRefunding        = 4
	MidRefunded         = 5
	SecondPending       = 6
	SecondRefunding     = 7
	SecondRefunded      = 8
	SecondSuccess       = 9
)

var StatusStr = map[int]string{
	SubmitFailed:    "Submit failed",
	FirstPending:    "Pending",
	FirstRefunding:  "Refunding",
	FirstRefunded:   "Refunded",
	MidRefunding:    "Refunding",
	MidRefunded:     "Refunded",
	SecondPending:   "Pending",
	SecondSuccess:   "Success",
	SecondRefunding: "Refunding",
	SecondRefunded:  "Refunded",
}

const INTERSWAP_TX_TOPIC = "interswaptx_topic"

// task by swap path
const InterswapPdexPappTxTask = "interswaptx_pathtype1" // Path 1: pDEX => pApp
const InterswapPappPdexTask = "interswaptx_pathtype2"   // Path 2: pApp => pDEX
