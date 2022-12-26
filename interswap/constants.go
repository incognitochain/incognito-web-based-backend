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

const (
	FirstPending    int = 1
	FirstSuccess        = 2
	FirstRefunding      = 3
	FirstRefunded       = 4
	SecondPending       = 5
	SecondSuccess       = 6
	SecondRefunding     = 7
	SecondRefunded      = 8
)

var StatusStr = map[int]string{
	FirstPending:    "Pending",
	FirstSuccess:    "Pending",
	FirstRefunding:  "Refunding",
	FirstRefunded:   "Refunded",
	SecondPending:   "Pending",
	SecondSuccess:   "Success",
	SecondRefunding: "Refunding",
	SecondRefunded:  "Refunded",
}

const INTERSWAP_TX_TOPIC = "interswaptx_topic"

// task by swap path
const InterswapPdexPappTxTask = "interswaptx_pathtype1" // Path 1: pDEX => pApp
const InterswapPappPdexTask = "interswaptx_pathtype2"   // Path 2: pApp => pDEX
