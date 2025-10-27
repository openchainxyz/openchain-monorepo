package ethclient

type Chain string

const (
	// etherscan
	EthereumMainnet   = "ethereum"
	EthereumRopsten   = "ropsten"
	EthereumKovan     = "kovan"
	EthereumRinkeby   = "rinkeby"
	EthereumGoerli    = "goerli"
	EthereumSepolia   = "sepolia"
	BinanceMainnet    = "binance"
	BinanceTestnet    = "binance-testnet"
	FantomMainnet     = "fantom"
	FantomTestnet     = "fantom-testnet"
	HecoMainnet       = "heco"
	HecoTestnet       = "heco-testnet"
	OptimismMainnet   = "optimism"
	OptimismTestnet   = "optimism-kovan"
	PolygonMainnet    = "polygon"
	PolygonTestnet    = "mumbai"
	ArbitrumMainnet   = "arbitrum"
	ArbitrumTestnet   = "arbitrum-testnet"
	HooMainnet        = "hoo"
	MoonbeamMainnet   = "moonbeam"
	MoonriverMainnet  = "moonriver"
	MoonbaseTestnet   = "moonbase"
	AvalancheMainnet  = "avalanche"
	AvalancheTestnet  = "avalanche-testnet"
	CronosMainnet     = "cronos"
	CronosTestnet     = "cronos-testnet"
	BittorrentMainnet = "bittorrent"
	BittorrentTestnet = "bittorrent-testnet"
	ZkSyncEraMainnet  = "zksync"
	LineaMainnet      = "linea"

	// blockscout
	POAMainnet   = "poa"
	POATestnet   = "poa-sokol"
	XDaiMainnet  = "xdai"
	XDaiArbitrum = "xdai-arbitrum"
	XDaiOptimism = "xdai-optimism"
	LuksoL14     = "lukso-l14"
	RskMainnet   = "rsk"
	EtcMainnet   = "etc"
	EtcKotti     = "etc-kotti"
	EtcMordor    = "etc-mordor"

	FlareMainnet = "flare-songbird"
	FlareTestnet = "flare-coston"

	BobaMainnet      = "boba"
	BobaRinkeby      = "boba-rinkeby"
	VelasMainnet     = "velas"
	EnergyWebMainnet = "energyweb"
	EnergyWebTestnet = "energyweb-volta"
	MetisTestnet     = "metis-stardust"
	MetisMainnet     = "metis-andromeda"
	MilkomedaMainnet = "milkomeda"
	CeloMainnet      = "celo"
	CeloAlfajores    = "celo-alfajores"
	CeloBaklava      = "celo-baklava"
	AuroraMainnet    = "aurora"
	AuroraTestnet    = "aurora-testnet"
	CLVMainnet       = "clv"
	FuseMainnet      = "fuse"
	FuseTestnet      = "fuse-spark"
	EnergiMainnet    = "energi"
	EnergiTestnet    = "energi-testnet"
	AstarMainnet     = "astar"
	AstarShiden      = "astar-shiden"
	AstarShibuya     = "astar-shibuya"

	// sourcify
	CandleMainnet           = "candle"
	DefiKingdomsMainnet     = "dfk"
	DefiKingdomsTestnet     = "dfk-testnet"
	DarwiniaCrabMainnet     = "crab"
	DarwiniaPangolinTestnet = "crab-testnet"
	EVMOSMainnet            = "evmos"
	EVMOSTestnet            = "evmos-testnet"
	GatherMainnet           = "gather"
	GatherTestnet           = "gather-testnet"
	GatherDevnet            = "gather-devnet"
	MeterMainnet            = "meter"
	MeterTestnet            = "meter-testnet"
	MultivacMainnet         = "multivac"
	OneLedgerMainnet        = "oneledger"
	OneLedgerTestnet        = "oneledger-frankenstein"
	PalmMainnet             = "palm"
	PalmTestnet             = "palm-testnet"
	SyscoinMainnet          = "syscoin"
	SyscoinTestnet          = "syscoin-testnet"
	TelosMainnet            = "telos"
	TelosTestnet            = "telos-testnet"
	UbiqMainnet             = "ubiq"
	WAGMIMainnet            = "wagmi"
)

var ChainIDs = map[Chain]int{
	EthereumMainnet:   1,
	EthereumRopsten:   3,
	EthereumRinkeby:   4,
	EthereumGoerli:    5,
	EtcKotti:          6,
	OptimismMainnet:   10,
	FlareTestnet:      16,
	FlareMainnet:      19,
	LuksoL14:          22,
	CronosMainnet:     25,
	BobaRinkeby:       28,
	RskMainnet:        30,
	EthereumKovan:     42,
	BinanceMainnet:    56,
	EtcMainnet:        61,
	EtcMordor:         63,
	OptimismTestnet:   69,
	HooMainnet:        70,
	POATestnet:        77,
	AstarShibuya:      81,
	BinanceTestnet:    97,
	POAMainnet:        99,
	XDaiMainnet:       100,
	VelasMainnet:      106,
	FuseMainnet:       122,
	FuseTestnet:       123,
	HecoMainnet:       128,
	PolygonMainnet:    137,
	BittorrentMainnet: 199,
	XDaiArbitrum:      200,
	EnergyWebMainnet:  246,
	FantomMainnet:     250,
	HecoTestnet:       256,
	BobaMainnet:       288,
	XDaiOptimism:      300,
	ZkSyncEraMainnet:  324,
	AstarShiden:       336,
	CronosTestnet:     338,
	MetisTestnet:      588,
	AstarMainnet:      592,
	BittorrentTestnet: 1028,
	MetisMainnet:      1088,
	CLVMainnet:        1024,
	MoonbeamMainnet:   1284,
	MoonriverMainnet:  1285,
	MoonbaseTestnet:   1287,
	MilkomedaMainnet:  2001,
	FantomTestnet:     4002,
	EnergiMainnet:     39797,
	ArbitrumMainnet:   42161,
	CeloMainnet:       42220,
	AvalancheTestnet:  43113,
	AvalancheMainnet:  43114,
	CeloAlfajores:     44787,
	EnergiTestnet:     49797,
	LineaMainnet:      59144,
	CeloBaklava:       62320,
	EnergyWebTestnet:  73799,
	PolygonTestnet:    80001,
	ArbitrumTestnet:   421611,
	EthereumSepolia:   11155111,
	AuroraMainnet:     1313161554,
	AuroraTestnet:     1313161555,

	CandleMainnet:           534,
	DefiKingdomsMainnet:     53935,
	DefiKingdomsTestnet:     335,
	DarwiniaCrabMainnet:     44,
	DarwiniaPangolinTestnet: 43,
	EVMOSMainnet:            9001,
	EVMOSTestnet:            9000,
	GatherDevnet:            486217935,
	GatherMainnet:           192837465,
	GatherTestnet:           356256156,
	MeterMainnet:            82,
	MeterTestnet:            83,
	MultivacMainnet:         62621,
	OneLedgerMainnet:        311752642,
	OneLedgerTestnet:        4216137055,
	PalmMainnet:             11297108109,
	PalmTestnet:             11297108099,
	SyscoinMainnet:          57,
	SyscoinTestnet:          5700,
	TelosMainnet:            40,
	TelosTestnet:            41,
	UbiqMainnet:             8,
	WAGMIMainnet:            11111,
}

var IDToChain = make(map[int]Chain)

func init() {
	for k, v := range ChainIDs {
		IDToChain[v] = k
	}
}
