package api

const (
	ProtocolFee = "protocol_fee"
	NetworkFee  = "network_fee"
)

const (
	cacheVaultStateKey           = "cache_vault_state"
	cacheSupportedPappsTokensKey = "cache_supported_papps_tokens"
	cacheTokenListKey            = "cache_token_list"
	cacheNetworkInfosKey         = "cache_network_infos"
	cacheCurvePoolIndexKey       = "cache_curve_pool_index"
)
const ethDefault string = `[
	{
		"id": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		"symbol": "WETH",
		"volumeUSD": "637023944489.7999978663192031278146"
	},
	{
		"id": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		"symbol": "USDC",
		"volumeUSD": "482078090643.2589753518605780720012"
	},
	{
		"id": "0xdac17f958d2ee523a2206206994597c13d831ec7",
		"symbol": "USDT",
		"volumeUSD": "130740647099.0354524438989385761747"
	},
	{
		"id": "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
		"symbol": "WBTC",
		"volumeUSD": "73868596320.7522880512219135805134"
	},
	{
		"id": "0x6b175474e89094c44da98b954eedeac495271d0f",
		"symbol": "DAI",
		"volumeUSD": "61978817943.03824136114149658845364"
	},
	{
		"id": "0x956f47f50a910163d8bf957cf5846d573e7f87ca",
		"symbol": "FEI",
		"volumeUSD": "10350473248.83117494326932401295114"
	},
	{
		"id": "0xa47c8bf37f92abed4a126bda807a7b7498661acd",
		"symbol": "UST",
		"volumeUSD": "7681963825.437804199162607843552691"
	},
	{
		"id": "0x4d224452801aced8b2f0aebe155379bb5d594381",
		"symbol": "APE",
		"volumeUSD": "7089812504.18024524867765117171013"
	},
	{
		"id": "0xf4d2888d29d722226fafa5d9b24f9164c092421e",
		"symbol": "LOOKS",
		"volumeUSD": "6356691923.836460731206913303746754"
	},
	{
		"id": "0x2b591e99afe9f32eaa6214f7b7629768c40eeb39",
		"symbol": "HEX",
		"volumeUSD": "5812616028.404744392657206513075794"
	},
	{
		"id": "0x853d955acef822db058eb8505911ed77f175b99e",
		"symbol": "FRAX",
		"volumeUSD": "5629534686.675541763232334300452646"
	},
	{
		"id": "0x514910771af9ca656af840dff83e8264ecf986ca",
		"symbol": "LINK",
		"volumeUSD": "4482775458.189329439637192843304396"
	},
	{
		"id": "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
		"symbol": "UNI",
		"volumeUSD": "4152604131.26697748679848332745027"
	},
	{
		"id": "0x95ad61b0a150d79219dcf64e1e6cc01f0b64c4ce",
		"symbol": "SHIB",
		"volumeUSD": "3583202359.900554027486649592481075"
	},
	{
		"id": "0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0",
		"symbol": "MATIC",
		"volumeUSD": "3340333410.11939971694301214672972"
	},
	{
		"id": "0xc18360217d8f7ab5e7c516566761ea12ce7f9d72",
		"symbol": "ENS",
		"volumeUSD": "2761015608.869778347502223175500701"
	},
	{
		"id": "0xaa6e8127831c9de45ae56bb1b0d4d4da6e5665bd",
		"symbol": "ETH2x-FLI",
		"volumeUSD": "2721560157.81689421849089630873565"
	},
	{
		"id": "0x92d6c1e31e14520e676a687f0a93788b716beff5",
		"symbol": "DYDX",
		"volumeUSD": "2456765637.905238969570517219317615"
	},
	{
		"id": "0x8e870d67f660d95d5be530380d0ec0bd388289e1",
		"symbol": "PAX",
		"volumeUSD": "2240256674.884211596275003410717704"
	},
	{
		"id": "0xbb0e17ef65f82ab018d8edd776e8dd940327b28b",
		"symbol": "AXS",
		"volumeUSD": "2209199614.754902395652003923157604"
	},
	{
		"id": "0xc581b735a1688071a1746c968e0798d642ede491",
		"symbol": "EURT",
		"volumeUSD": "1776915529.262136928920673503710603"
	},
	{
		"id": "0x15d4c048f83bd7e37d49ea4c83a07267ec4203da",
		"symbol": "GALA",
		"volumeUSD": "1490033212.312540515303030997028105"
	},
	{
		"id": "0x9e32b13ce7f2e80a01932b42553652e053d6ed8e",
		"symbol": "Metis",
		"volumeUSD": "1461595756.932811115938465250386097"
	},
	{
		"id": "0x1a7e4e63778b4f12a199c062f3efdd288afcbce8",
		"symbol": "agEUR",
		"volumeUSD": "1391257947.764908652175663333724156"
	},
	{
		"id": "0x99d8a9c45b2eca8864373a26d1459e3dff1e17f3",
		"symbol": "MIM",
		"volumeUSD": "1349662767.65224596416540726885254"
	},
	{
		"id": "0x111111111117dc0aa78b770fa6a738034120c302",
		"symbol": "1INCH",
		"volumeUSD": "1339366256.637246980190373881389171"
	},
	{
		"id": "0x5f98805a4e8be255a32880fdec7f6728c6568ba0",
		"symbol": "LUSD",
		"volumeUSD": "1299544367.565339731147438081356054"
	},
	{
		"id": "0x03ab458634910aad20ef5f1c8ee96f1d6ac54919",
		"symbol": "RAI",
		"volumeUSD": "1282300000.034276635406676815546496"
	},
	{
		"id": "0x4e15361fd6b4bb609fa63c81a2be19d873717870",
		"symbol": "FTM",
		"volumeUSD": "1282200807.301727047465688424818192"
	},
	{
		"id": "0x6123b0049f904d730db3c36a31167d9d4121fa6b",
		"symbol": "RBN",
		"volumeUSD": "1227617478.589787804473610680456024"
	},
	{
		"id": "0xa693b19d2931d498c5b318df961919bb4aee87a5",
		"symbol": "UST",
		"volumeUSD": "1185914935.792825318474165676594689"
	},
	{
		"id": "0xeb4c2781e4eba804ce9a9803c67d0893436bb27d",
		"symbol": "renBTC",
		"volumeUSD": "1173800636.129973456724005558607218"
	},
	{
		"id": "0x090185f2135308bad17527004364ebcc2d37e5f6",
		"symbol": "SPELL",
		"volumeUSD": "1148831215.456878182521440747169277"
	},
	{
		"id": "0x32353a6c91143bfd6c7d363b546e62a9a2489a20",
		"symbol": "AGLD",
		"volumeUSD": "1108884490.984007671223061844509236"
	},
	{
		"id": "0x990f341946a3fdb507ae7e52d17851b87168017c",
		"symbol": "STRONG",
		"volumeUSD": "1094340946.887643337886634626462362"
	},
	{
		"id": "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2",
		"symbol": "MKR",
		"volumeUSD": "1093928424.559574984049674368399022"
	},
	{
		"id": "0xdb25f211ab05b1c97d595516f45794528a807ad8",
		"symbol": "EURS",
		"volumeUSD": "1029265639.679867413069160432968193"
	},
	{
		"id": "0x7a58c0be72be218b41c608b7fe7c5bb630736c71",
		"symbol": "PEOPLE",
		"volumeUSD": "972169238.5070026006314881590868009"
	},
	{
		"id": "0xc0d4ceb216b3ba9c3701b291766fdcba977cec3a",
		"symbol": "BTRFLY",
		"volumeUSD": "911337519.221175716182096536830901"
	},
	{
		"id": "0xaaaebe6fe48e54f431b0c390cfaf0b017d09d42d",
		"symbol": "CEL",
		"volumeUSD": "901093909.365673812081931051058612"
	},
	{
		"id": "0xcc8fa225d80b9c7d42f96e9570156c65d6caaa25",
		"symbol": "SLP",
		"volumeUSD": "891234006.8579928718663560696458232"
	},
	{
		"id": "0x4a220e6096b25eadb88358cb44068a3248254675",
		"symbol": "QNT",
		"volumeUSD": "880363620.6058343472382011266822704"
	},
	{
		"id": "0x419d0d8bdd9af5e606ae2232ed285aff190e711b",
		"symbol": "FUN",
		"volumeUSD": "841836017.0426719753603787619844836"
	},
	{
		"id": "0x3845badade8e6dff049820680d1f14bd3903a5d0",
		"symbol": "SAND",
		"volumeUSD": "826965792.7195468102714613121924967"
	},
	{
		"id": "0xd533a949740bb3306d119cc777fa900ba034cd52",
		"symbol": "CRV",
		"volumeUSD": "822033151.3674545866654769014130856"
	},
	{
		"id": "0x5a98fcbea516cf06857215779fd812ca3bef1b32",
		"symbol": "LDO",
		"volumeUSD": "809102648.6566247431466724225815654"
	},
	{
		"id": "0x4fabb145d64652a948d72533023f6e7a623c7c53",
		"symbol": "BUSD",
		"volumeUSD": "783125752.9962197272444019379486044"
	},
	{
		"id": "0x6b4c7a5e3f0b99fcd83e9c089bddd6c7fce5c611",
		"symbol": "MM",
		"volumeUSD": "770882015.9267559001147166440488281"
	},
	{
		"id": "0xb62132e35a6c13ee1ee0f84dc5d40bad8d815206",
		"symbol": "NEXO",
		"volumeUSD": "768217196.7235941734245243514164654"
	},
	{
		"id": "0xba5bde662c17e2adff1075610382b9b691296350",
		"symbol": "RARE",
		"volumeUSD": "763827366.2490002138863326348524216"
	}
]
`
const plgDefault string = `[
	{
		"id": "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
		"symbol": "USDC",
		"volumeUSD": "19287714337.20098606205566802265652"
	},
	{
		"id": "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619",
		"symbol": "WETH",
		"volumeUSD": "16950137882.99740220482414485575389"
	},
	{
		"id": "0x0d500b1d8e8ef31e21c99d1db9a6444d3adf1270",
		"symbol": "WMATIC",
		"volumeUSD": "9876686550.781954042456697413127704"
	},
	{
		"id": "0x1bfd67037b42cf73acf2047067bd4f2c47d9bfd6",
		"symbol": "WBTC",
		"volumeUSD": "2356581215.567670186229312571280106"
	},
	{
		"id": "0xc2132d05d31c914a87c6611c10748aeb04b58e8f",
		"symbol": "USDT",
		"volumeUSD": "1271379973.860127405298699679265644"
	},
	{
		"id": "0x8f3cf7ad23cd3cadbd9735aff958023239c6a063",
		"symbol": "DAI",
		"volumeUSD": "449145964.4287832902206513229872511"
	},
	{
		"id": "0xa3fa99a148fa48d14ed51d610c367c61876997f1",
		"symbol": "miMATIC",
		"volumeUSD": "135947129.7708380220050417552815954"
	},
	{
		"id": "0x3a58a54c066fdc0f2d55fc9c89f0415c92ebf3c4",
		"symbol": "stMATIC",
		"volumeUSD": "102260231.1473479369016450163781432"
	},
	{
		"id": "0x53e0bca35ec356bd5dddfebbd1fc0fd03fabad39",
		"symbol": "LINK",
		"volumeUSD": "68623144.52021634373349720509152133"
	},
	{
		"id": "0x45c32fa6df82ead1e2ef74d17b76547eddfaff89",
		"symbol": "FRAX",
		"volumeUSD": "60447101.42446031273172440132701971"
	},
	{
		"id": "0x172370d5cd63279efa6d502dab29171933a610af",
		"symbol": "CRV",
		"volumeUSD": "52850903.59669158211348917477096427"
	},
	{
		"id": "0xe61b839e87ffe2addd41b33f1b048f40fbb6a7f6",
		"symbol": "ODOGE",
		"volumeUSD": "41008402.10376649779574819179955544"
	},
	{
		"id": "0xd6df932a45c0f255f85145f286ea0b292b21c90b",
		"symbol": "AAVE",
		"volumeUSD": "40553165.546366379790694552630209"
	},
	{
		"id": "0xb33eaad8d922b1083446dc23f610c2567fb5180f",
		"symbol": "UNI",
		"volumeUSD": "28811880.54868277123634468534926486"
	},
	{
		"id": "0x0d0b8488222f7f83b23e365320a4021b12ead608",
		"symbol": "NXTT",
		"volumeUSD": "25860606.09991208450693618515306698"
	},
	{
		"id": "0x2760e46d9bb43dafcbecaad1f64b93207f9f0ed7",
		"symbol": "MVX",
		"volumeUSD": "18625490.54326071752947456585973522"
	},
	{
		"id": "0xe0b52e49357fd4daf2c15e02058dce6bc0057db4",
		"symbol": "agEUR",
		"volumeUSD": "13892957.5071564887979997553933805"
	},
	{
		"id": "0x235737dbb56e8517391473f7c964db31fa6ef280",
		"symbol": "KASTA",
		"volumeUSD": "12847785.77722743417913012230754084"
	},
	{
		"id": "0xe2aa7db6da1dae97c5f5c6914d285fbfcc32a128",
		"symbol": "PAR",
		"volumeUSD": "9561614.749034"
	},
	{
		"id": "0x301595f6fd5f69fad7a488dacb8971e7c0c2f559",
		"symbol": "wtPOKT",
		"volumeUSD": "8726610.117545129260941514659057037"
	},
	{
		"id": "0xdc3326e71d45186f113a2f448984ca0e8d201995",
		"symbol": "XSGD",
		"volumeUSD": "7821311.175152503573270798171225243"
	},
	{
		"id": "0xe5417af564e4bfda1c483642db72007871397896",
		"symbol": "GNS",
		"volumeUSD": "7756163.443744952631477364959253446"
	},
	{
		"id": "0x2ab4f9ac80f33071211729e45cfc346c1f8446d5",
		"symbol": "CGG",
		"volumeUSD": "5485463.579984152118630240642069066"
	},
	{
		"id": "0xacd4e2d936be9b16c01848a3742a34b3d5a5bdfa",
		"symbol": "$MECHA",
		"volumeUSD": "4071719.553382776595240318896703213"
	},
	{
		"id": "0x0e2c818fea38e7df50410f772b7d59af20589a62",
		"symbol": "DOM",
		"volumeUSD": "3583972.955727958467814756177339354"
	},
	{
		"id": "0xb0b195aefa3650a6908f15cdac7d92f8a5791b0b",
		"symbol": "BOB",
		"volumeUSD": "3097754.837380593745226196443245557"
	},
	{
		"id": "0xc3c7d422809852031b44ab29eec9f1eff2a58756",
		"symbol": "LDO",
		"volumeUSD": "2950754.430374520855862428375580131"
	},
	{
		"id": "0xbbba073c31bf03b8acf7c28ef0738decf3695683",
		"symbol": "SAND",
		"volumeUSD": "2768317.708902603294093218123279118"
	},
	{
		"id": "0x35b51ff33be10a9a741e9c9d3f17585e4b7d15c0",
		"symbol": "indexUSDC",
		"volumeUSD": "2764993.110268"
	},
	{
		"id": "0xa9f37d84c856fda3812ad0519dad44fa0a3fe207",
		"symbol": "MLN",
		"volumeUSD": "2184788.384842609682369937100662193"
	},
	{
		"id": "0xadbe0eac80f955363f4ff47b0f70189093908c04",
		"symbol": "XMT",
		"volumeUSD": "2168046.87992609367432616653519747"
	},
	{
		"id": "0x486ffaf06a681bf22b5209e9ffce722662a60e8c",
		"symbol": "FLY",
		"volumeUSD": "1885155.700912"
	},
	{
		"id": "0x62a872d9977db171d9e213a5dc2b782e72ca0033",
		"symbol": "NEUY",
		"volumeUSD": "1495135.15474099335064346221071221"
	},
	{
		"id": "0x692c44990e4f408ba0917f5c78a83160c1557237",
		"symbol": "THALES",
		"volumeUSD": "1296116.73944"
	},
	{
		"id": "0x30de46509dbc3a491128f97be0aaf70dc7ff33cb",
		"symbol": "XZAR",
		"volumeUSD": "1004937.004411158135884090169372599"
	},
	{
		"id": "0xe111178a87a3bff0c8d18decba5798827539ae99",
		"symbol": "EURS",
		"volumeUSD": "907202.6021977093602731162344281718"
	},
	{
		"id": "0x111111517e4929d3dcbdfa7cce55d30d4b6bc4d6",
		"symbol": "ICHI",
		"volumeUSD": "704289.6847617969605452911355963603"
	},
	{
		"id": "0xed755dba6ec1eb520076cec051a582a6d81a8253",
		"symbol": "CHAMP",
		"volumeUSD": "614559.7396875409564713496806803833"
	},
	{
		"id": "0xc5b57e9a1e7914fda753a88f24e5703e617ee50c",
		"symbol": "POP",
		"volumeUSD": "579193.805947"
	},
	{
		"id": "0x09a84f900205b1ac5f3214d3220c7317fd5f5b77",
		"symbol": "FREC",
		"volumeUSD": "549668.0428887887232178796998870244"
	},
	{
		"id": "0x2934b36ca9a4b31e633c5be670c8c8b28b6aa015",
		"symbol": "THX",
		"volumeUSD": "504482.157023"
	},
	{
		"id": "0xdfce1e99a31c4597a3f8a8945cbfa9037655e335",
		"symbol": "ASTRAFER",
		"volumeUSD": "304528.9722504469263062917922439058"
	},
	{
		"id": "0x2c826035c1c36986117a0e949bd6ad4bab54afe2",
		"symbol": "XIDR",
		"volumeUSD": "256193.176377"
	},
	{
		"id": "0xc75ea0c71023c14952f3c7b9101ecbbaa14aa27a",
		"symbol": "NFTI",
		"volumeUSD": "254019.0872256712003239743523778148"
	},
	{
		"id": "0x9c9e5fd8bbc25984b178fdce6117defa39d2db39",
		"symbol": "BUSD",
		"volumeUSD": "235221.395731238686081594301488217"
	},
	{
		"id": "0xc145718228438a045d76d11248fb779e4d23f942",
		"symbol": "Zi",
		"volumeUSD": "110780.817806"
	},
	{
		"id": "0x0e9b89007eee9c958c0eda24ef70723c2c93dd58",
		"symbol": "aMATICc",
		"volumeUSD": "65445.47319697084874057924544291182"
	},
	{
		"id": "0x8d52c2d70a7c28a9daac2ff12ad9bfbf041cd318",
		"symbol": "CIAO",
		"volumeUSD": "42077.32160562249354508360774300325"
	},
	{
		"id": "0xe0d4a49c386f5c0184f72ac0752aad4bd62c579a",
		"symbol": "FTXXX",
		"volumeUSD": "2810.095822450660991653079439150723"
	},
	{
		"id": "0x9a5c2f40910b3d0e97defab7d775cd408085c14e",
		"symbol": "WRAC",
		"volumeUSD": "1546.413173503516033080449015245035"
	}
]
`
const ftmDefault string = `[
	{
		"id": "0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83",
		"symbol": "WFTM",
		"tradeVolumeUSD": "57298836420.40470509894393107966719"
	},
	{
		"id": "0x04068da6c83afcfa0e13ba15a6696662335d5b75",
		"symbol": "USDC",
		"tradeVolumeUSD": "23065583952.19498840758325960066126"
	},
	{
		"id": "0x8d11ec38a3eb5e956b052f67da8bdc9bef8abf3e",
		"symbol": "DAI",
		"tradeVolumeUSD": "6576773532.326554882512030929040167"
	},
	{
		"id": "0x049d68029688eabf473097a2fc38ef61633a3c7a",
		"symbol": "fUSDT",
		"tradeVolumeUSD": "4615613817.236313508920854923628646"
	},
	{
		"id": "0x4cdf39285d7ca8eb3f090fda0c069ba5f4145b37",
		"symbol": "TSHARE",
		"tradeVolumeUSD": "3824421705.617060909725143400141188"
	},
	{
		"id": "0x6c021ae822bea943b2e66552bde1d2696a53fbb7",
		"symbol": "TOMB",
		"tradeVolumeUSD": "3668508255.335313597711454802153178"
	},
	{
		"id": "0x74b23882a30290451a17c44f4f05243b6b58c76d",
		"symbol": "ETH",
		"tradeVolumeUSD": "3543257405.722492556756435453951163"
	},
	{
		"id": "0x82f0b8b456c1a451378467398982d4834b6829c1",
		"symbol": "MIM",
		"tradeVolumeUSD": "2081180693.063564306982401953135522"
	},
	{
		"id": "0x321162cd933e2be498cd2267a90534a804051b11",
		"symbol": "BTC",
		"tradeVolumeUSD": "1927790363.859521115830973153011053"
	},
	{
		"id": "0x841fad6eae12c286d1fd18d1d525dffa75c7effe",
		"symbol": "BOO",
		"tradeVolumeUSD": "1524394923.423570896021692355196538"
	},
	{
		"id": "0xc54a1684fd1bef1f077a336e6be4bd9a3096a6ca",
		"symbol": "2SHARES",
		"tradeVolumeUSD": "1076044144.609394266967244171575425"
	},
	{
		"id": "0xd67de0e0a0fd7b15dc8348bb9be742f3c5850454",
		"symbol": "BNB",
		"tradeVolumeUSD": "1057692698.264904771350101872837359"
	},
	{
		"id": "0x5c4fdfc5233f935f20d2adba572f770c2e377ab0",
		"symbol": "HEC",
		"tradeVolumeUSD": "1008821859.000056830831718181397586"
	},
	{
		"id": "0xf16e81dce15b08f326220742020379b855b87df9",
		"symbol": "ICE",
		"tradeVolumeUSD": "785184927.8898364297404539269152187"
	},
	{
		"id": "0xfb98b335551a418cd0737375a2ea0ded62ea213b",
		"symbol": "miMATIC",
		"tradeVolumeUSD": "779069507.6087888105746423600660931"
	},
	{
		"id": "0xc165d941481e68696f43ee6e99bfb2b23e0e3114",
		"symbol": "OXD",
		"tradeVolumeUSD": "649150039.8800525428613541579240064"
	},
	{
		"id": "0xe0654c8e6fd4d733349ac7e09f6f23da256bf475",
		"symbol": "SCREAM",
		"tradeVolumeUSD": "579220285.6883318602538207886638055"
	},
	{
		"id": "0x49c290ff692149a4e16611c694fded42c954ab7a",
		"symbol": "BSHARE",
		"tradeVolumeUSD": "560050397.9418503818646003301694251"
	},
	{
		"id": "0x5602df4a94eb6c680190accfa2a475621e0ddbdc",
		"symbol": "SPA",
		"tradeVolumeUSD": "519909943.8821523398988953328520086"
	},
	{
		"id": "0x468003b688943977e6130f4f68f23aad939a1040",
		"symbol": "SPELL",
		"tradeVolumeUSD": "474370602.6591893706035538065610638"
	}
]
`
const bscDefault string = `[{
	"id": "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
	"hydeSymbol": "WBNB",
	"tradeVolumeUSD": 550263357.4604223
},
{
	"id": "0xe9e7cea3dedca5984780bafc599bd69add087d56",
	"hydeSymbol": "BUSD",
	"tradeVolumeUSD": 250505203.55710915
},
{
	"id": "0x55d398326f99059ff775485246999027b3197955",
	"hydeSymbol": "USDT",
	"tradeVolumeUSD": 207062430.39510247
},
{
	"id": "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d",
	"hydeSymbol": "USDC",
	"tradeVolumeUSD": 54979178.067721024
},
{
	"id": "0xb82beb6ee0063abd5fc8e544c852237aa62cbb14",
	"hydeSymbol": "SQUA",
	"tradeVolumeUSD": 49701478.424213685
},
{
	"id": "0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c",
	"hydeSymbol": "BTCB",
	"tradeVolumeUSD": 43429230.22033902
},
{
	"id": "0x2170ed0880ac9a755fd29b2688956bd959f933f8",
	"hydeSymbol": "ETH",
	"tradeVolumeUSD": 42557743.0515794
},
{
	"id": "0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82",
	"hydeSymbol": "Cake",
	"tradeVolumeUSD": 33940611.88009659
},
{
	"id": "0xf9324d2477072278aed4a84d7163f51e06fe7ead",
	"hydeSymbol": "GALA",
	"tradeVolumeUSD": 12118663.55351225
},
{
	"id": "0x26619fa1d4c957c58096bbbeca6588dcfb12e109",
	"hydeSymbol": "TIME",
	"tradeVolumeUSD": 9568139.414233938
},
{
	"id": "0xbfd48cc239bc7e7cd5ad9f9630319f9b59e0b9e1",
	"hydeSymbol": "CCDS",
	"tradeVolumeUSD": 6970430.98617727
},
{
	"id": "0x109b451f0a724e7a3d99bbac1aa5d58a4821a5b0",
	"hydeSymbol": "FTT",
	"tradeVolumeUSD": 6958344.238341204
},
{
	"id": "0xf8a0bf9cf54bb92f17374d9e9a321e6a111a51bd",
	"hydeSymbol": "LINK",
	"tradeVolumeUSD": 5255590.557177101
},
{
	"id": "0x198271b868dae875bfea6e6e4045cdda5d6b9829",
	"hydeSymbol": "AFD",
	"tradeVolumeUSD": 4802550.201070539
},
{
	"id": "0x9244a9836ba889b970293e8eb958c949902dce1b",
	"hydeSymbol": "sDoge",
	"tradeVolumeUSD": 4440646.828610331
},
{
	"id": "0x85eac5ac2f758618dfa09bdbe0cf174e7d574d5b",
	"hydeSymbol": "TRX",
	"tradeVolumeUSD": 4081440.9042740623
},
{
	"id": "0xc632f90affec7121120275610bf17df9963f181c",
	"hydeSymbol": "DEBT",
	"tradeVolumeUSD": 4014912.272964508
},
{
	"id": "0x277ae79c42c859ca858d5a92c22222c8b65c6d94",
	"hydeSymbol": "ABB",
	"tradeVolumeUSD": 3869804.85629125
},
{
	"id": "0x8c851d1a123ff703bd1f9dabe631b69902df5f97",
	"hydeSymbol": "BNX",
	"tradeVolumeUSD": 3821748.267749726
},
{
	"id": "0x1d2f0da169ceb9fc7b3144628db156f3f6c60dbe",
	"hydeSymbol": "XRP",
	"tradeVolumeUSD": 3765680.9209038187
},
{
	"id": "0x8a87c36bb9e9b91c76e7a0a374a59e57cf0c0f5b",
	"hydeSymbol": "SUC",
	"tradeVolumeUSD": 3329965.1988722333
},
{
	"id": "0x3ee2200efb3400fabb9aacf31297cbdd1d435d47",
	"hydeSymbol": "ADA",
	"tradeVolumeUSD": 3213138.739570573
},
{
	"id": "0xba2ae424d960c26247dd6c32edc70b295c744c43",
	"hydeSymbol": "DOGE",
	"tradeVolumeUSD": 3114341.7084237337
},
{
	"id": "0xad6742a35fb341a9cc6ad674738dd8da98b94fb1",
	"hydeSymbol": "WOM",
	"tradeVolumeUSD": 3068634.7622439363
},
{
	"id": "0x7083609fce4d1d8dc0c979aab8c869ea2c873402",
	"hydeSymbol": "DOT",
	"tradeVolumeUSD": 2970095.5989784864
},
{
	"id": "0x7ddee176f665cd201f93eede625770e2fd911990",
	"hydeSymbol": "GALA",
	"tradeVolumeUSD": 2726849.0260506854
},
{
	"id": "0x4b0f1812e5df2a09796481ff14017e6005508003",
	"hydeSymbol": "TWT",
	"tradeVolumeUSD": 2687075.2575622755
},
{
	"id": "0x3203c9e46ca618c8c1ce5dc67e7e9d75f5da2377",
	"hydeSymbol": "MBOX",
	"tradeVolumeUSD": 2601535.483676543
},
{
	"id": "0x8bf9dc93b6f81a5fc70d0b451596fd2b09fe92c3",
	"hydeSymbol": "TAU",
	"tradeVolumeUSD": 2488277.766350165
},
{
	"id": "0xd41fdb03ba84762dd66a0af1a6c8540ff1ba5dfb",
	"hydeSymbol": "SFP",
	"tradeVolumeUSD": 2401809.6699299044
},
{
	"id": "0xa045e37a0d1dd3a45fefb8803d22457abc0a728a",
	"hydeSymbol": "GHNY",
	"tradeVolumeUSD": 2396620.9016439007
},
{
	"id": "0x3019bf2a2ef8040c242c9a4c5c4bd4c81678b2a1",
	"hydeSymbol": "GMT",
	"tradeVolumeUSD": 2168501.954287144
},
{
	"id": "0x256d1fce1b1221e8398f65f9b36033ce50b2d497",
	"hydeSymbol": "wALV",
	"tradeVolumeUSD": 2060244.228772994
},
{
	"id": "0x12bb890508c125661e03b09ec06e404bc9289040",
	"hydeSymbol": "RACA",
	"tradeVolumeUSD": 1946443.6746264896
},
{
	"id": "0x156ab3346823b651294766e23e6cf87254d68962",
	"hydeSymbol": "LUNA",
	"tradeVolumeUSD": 1917877.813890785
},
{
	"id": "0xa123ab52a32267dc357b7599739d3c6caf856fe4",
	"hydeSymbol": "AIR",
	"tradeVolumeUSD": 1831701.9345290428
},
{
	"id": "0x3b5e381130673f794a5cf67fbba48688386bea86",
	"hydeSymbol": "POT",
	"tradeVolumeUSD": 1827581.532985718
},
{
	"id": "0xbf5140a22578168fd562dccf235e5d43a02ce9b1",
	"hydeSymbol": "UNI",
	"tradeVolumeUSD": 1777010.9365067952
},
{
	"id": "0x4803ac6b79f9582f69c4fa23c72cb76dd1e46d8d",
	"hydeSymbol": "TMT",
	"tradeVolumeUSD": 1494008.0707355866
},
{
	"id": "0x60322971a672b81bcce5947706d22c19daecf6fb",
	"hydeSymbol": "MDAO",
	"tradeVolumeUSD": 1453028.3823668004
},
{
	"id": "0xb0e384b53cfdc4417e66d5c74e955c3926b19c78",
	"hydeSymbol": "MATRIX",
	"tradeVolumeUSD": 1444090.9559675343
},
{
	"id": "0xf750a26eb0acf95556e8529e72ed530f3b60f348",
	"hydeSymbol": "GNT",
	"tradeVolumeUSD": 1421833.6799951852
},
{
	"id": "0x697bd938e7e572e787ecd7bc74a31f1814c21264",
	"hydeSymbol": "DIFX",
	"tradeVolumeUSD": 1414368.5848547316
},
{
	"id": "0xd66b7fec3f891f8a732e489c4591d5e2c4303091",
	"hydeSymbol": "MC",
	"tradeVolumeUSD": 1344637.6407116903
},
{
	"id": "0xc836d8dc361e44dbe64c4862d55ba041f88ddd39",
	"hydeSymbol": "WMATIC",
	"tradeVolumeUSD": 1237963.1705227373
},
{
	"id": "0xe283d0e3b8c102badf5e8166b73e02d96d92f688",
	"hydeSymbol": "ELEPHANT",
	"tradeVolumeUSD": 1165970.7633882188
},
{
	"id": "0xcf6bb5389c92bdda8a3747ddb454cb7a64626c63",
	"hydeSymbol": "XVS",
	"tradeVolumeUSD": 1098712.1248981827
},
{
	"id": "0xa08d56a5a7f61dfde27506ea8750a31bd0df65ae",
	"hydeSymbol": "W3U",
	"tradeVolumeUSD": 1096936.4717027904
},
{
	"id": "0x8f0528ce5ef7b51152a59745befdd91d97091d2f",
	"hydeSymbol": "ALPACA",
	"tradeVolumeUSD": 1082268.0590452938
}
]`
