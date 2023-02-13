import { doApiRequest, TraceEntry, TraceResponse } from '@components/tracer/api';
import { ChainConfig, ChainConfigContext, defaultChainConfig, getChain } from '@components/tracer/Chains';
import { DecodeTree } from '@components/tracer/decoder-format/DecodeTree';
import { GasPriceEstimator } from '@components/tracer/gas-price-estimator/estimate';
import { defaultLabelMetadata, LabelMetadata, LabelMetadataContext } from '@components/tracer/metadata/labels';
import {
    defaultPreimageMetadata,
    PreimageMetadata,
    PreimageMetadataContext,
} from '@components/tracer/metadata/preimages';
import {
    defaultPriceMetadata,
    fetchDefiLlamaPrices,
    PriceMetadata,
    PriceMetadataContext,
} from '@components/tracer/metadata/prices';
import { defaultTokenMetadata, TokenMetadata, TokenMetadataContext } from '@components/tracer/metadata/tokens';
import { TransactionMetadata, TransactionMetadataContext } from '@components/tracer/metadata/transaction';
import { precompiles } from '@components/tracer/precompiles';
import { TraceTree } from '@components/tracer/trace/TraceTree';
import { TransactionInfo } from '@components/tracer/transaction-info/TransactionInfo';
import { Result, TraceMetadata } from '@components/tracer/types';
import { ValueChange } from '@components/tracer/value-change/ValueChange';
import { Box, Typography } from '@mui/material';
import { Interface, JsonRpcProvider, Provider, TransactionReceipt } from 'ethers';
import { useRouter } from 'next/router';
import * as React from 'react';
import Navbar from '../../../components/Navbar';

export default function TransactionViewer() {
    const router = useRouter();
    const { chain, txhash } = router.query;

    const [chainConfig, setChainConfig] = React.useState<ChainConfig>(defaultChainConfig());
    const [provider, setProvider] = React.useState<Provider>();
    const [estimator, setEstimator] = React.useState<GasPriceEstimator>();

    const [transactionMetadata, setTransactionMetadata] = React.useState<Result<TransactionMetadata>>();

    const [traceResponse, setTraceResponse] = React.useState<Result<TraceResponse>>();

    const [preimageMetadata, setPreimageMetadata] = React.useState<PreimageMetadata>(defaultPreimageMetadata());
    const [labelMetadata, setLabelMetadata] = React.useState<LabelMetadata>(defaultLabelMetadata());
    const [priceMetadata, setPriceMetadata] = React.useState<PriceMetadata>(defaultPriceMetadata());
    const [tokenMetadata, setTokenMetadata] = React.useState<TokenMetadata>(defaultTokenMetadata());

    const [traceResult, setTraceResult] = React.useState<TraceResponse>();
    const [traceMetadata, setTraceMetadata] = React.useState<TraceMetadata>();

    React.useMemo(async () => {
        if (!chain || Array.isArray(chain)) return;
        if (!txhash || Array.isArray(txhash)) return;

        const chainConfig = await getChain(chain);
        if (!chainConfig) return;

        setChainConfig(chainConfig);

        setTokenMetadata({
            ...defaultTokenMetadata(),
            updater: setTokenMetadata,
        });
        setPriceMetadata({
            ...defaultPriceMetadata(),
            updater: setPriceMetadata,
        });
        setPreimageMetadata({
            ...defaultPreimageMetadata(),
            updater: setPreimageMetadata,
        });
        setTraceResult(undefined);
        setTransactionMetadata(undefined);

        const provider = new JsonRpcProvider(chainConfig.rpcUrl);
        setProvider(provider);

        const estimator = new GasPriceEstimator(provider);
        setEstimator(estimator);

        provider.getBlockNumber().catch(() => {});

        const tryFetchTrace = () => {
            doApiRequest<TraceResponse>(`/api/v1/trace/${chain}/${txhash}`)
                .then((traceResponse) => {
                    console.log('loaded trace', traceResponse);

                    let labels: Record<string, string> = {};
                    let customLabels: Record<string, Record<string, string>> = {};
                    try {
                        customLabels = JSON.parse(localStorage.getItem('pref:labels') || '{}');
                    } catch {}
                    if (!(chain in customLabels)) {
                        customLabels[chain] = {};
                    }

                    for (let address of Object.keys(precompiles)) {
                        labels[address] = 'Precompile';
                    }

                    let metadata: TraceMetadata = {
                        abis: {},
                        nodesByPath: {},
                    };

                    let preprocess = (node: TraceEntry) => {
                        metadata.nodesByPath[node.path] = node;

                        if (node.type === 'call') {
                            node.children.forEach(preprocess);
                        }
                    };
                    preprocess(traceResponse.entrypoint);

                    for (let [address, entries] of Object.entries(traceResponse.addresses)) {
                        metadata.abis[address] = {};
                        for (let [codehash, info] of Object.entries(entries)) {
                            labels[address] = labels[address] || info.label;

                            try {
                                console.log(info);
                                metadata.abis[address][codehash] = new Interface([
                                    ...Object.values(info.functions),
                                    ...Object.values(info.events),
                                    ...Object.values(info.errors).filter(
                                        (v) =>
                                            !(
                                                // lmao wtf ethers
                                                (
                                                    (v.name === 'Error' &&
                                                        v.inputs &&
                                                        v.inputs.length === 1 &&
                                                        v.inputs[0].type === 'string') ||
                                                    (v.name === 'Panic' &&
                                                        v.inputs &&
                                                        v.inputs.length === 1 &&
                                                        v.inputs[0].type === 'uint256')
                                                )
                                            ),
                                    ),
                                ]);
                            } catch (e) {
                                console.log('failed to construct interface', e);
                            }
                        }
                    }

                    for (let address of Object.keys(labels)) {
                        if (labels[address] === 'Vyper_contract') {
                            labels[address] = `Vyper_contract (0x${address.substring(2, 6)}..${address.substring(
                                38,
                                42,
                            )})`;
                        }
                    }

                    Object.keys(labels).forEach((addr) => delete customLabels[chain][addr]);
                    localStorage.setItem('pref:labels', JSON.stringify(customLabels));

                    setTraceResult(traceResponse);
                    setTraceMetadata(metadata);
                    setLabelMetadata({
                        updater: setLabelMetadata,
                        labels: labels,
                        customLabels: customLabels,
                    });
                    setTraceResponse({
                        ok: true,
                        result: traceResponse,
                    });
                })
                .catch((e) => {
                    setTraceResponse({
                        ok: false,
                        error: e,
                    });
                    console.log('failed to fetch trace', e);
                });
        };

        tryFetchTrace();

        Promise.allSettled([
            provider.getBlockNumber(), // make ethers fetch this so it gets batched (getTransactionReceipt really wants to know the confirmations)
            provider.getTransaction(txhash),
            provider.getTransactionReceipt(txhash),
        ]).then(([number, transactionResult, receiptResult]) => {
            if (number.status === 'rejected') {
                return;
            }
            if (transactionResult.status === 'rejected') {
                console.log('an error occurred while loading the transaction!', transactionResult.reason);

                setTransactionMetadata({
                    ok: false,
                    error: transactionResult.reason,
                });
                return;
            }

            if (!transactionResult.value) {
                setTransactionMetadata({
                    ok: false,
                    error: new Error('transaction not found'),
                });
                return;
            }

            const result: TransactionMetadata = {
                transaction: transactionResult.value,
                result: null,
            };

            const processReceipt = (receipt: TransactionReceipt) => {
                console.log('got receipt', receipt);
                result.result = {
                    receipt: receipt,
                    timestamp: Math.floor(new Date().getTime() / 1000),
                };

                setTransactionMetadata({
                    ok: true,
                    result: result,
                });

                if (number.value - receipt.blockNumber + 1 > 2) {
                    provider
                        .getBlock(receipt.blockHash)
                        .then((block) => {
                            result.result!.timestamp = block!.timestamp;

                            setTransactionMetadata({
                                ok: true,
                                result: result,
                            });

                            fetchDefiLlamaPrices(setPriceMetadata, [chainConfig.coingeckoId], block!.timestamp).catch(
                                (e) => {
                                    console.log('failed to fetch price', e);
                                },
                            );
                        })
                        .catch(() => {});
                }

                tryFetchTrace();
            };

            if (receiptResult.status === 'fulfilled' && receiptResult.value) {
                processReceipt(receiptResult.value);
            } else {
                provider
                    .waitForTransaction(txhash)
                    .then((r) => processReceipt(r!))
                    .catch((e) => {
                        console.log('error while waiting for receipt', e);
                    });
            }

            setTransactionMetadata({
                ok: true,
                result: result,
            });
        });
    }, [chain, txhash]);

    let transactionInfoGrid;
    if (transactionMetadata) {
        if (transactionMetadata.ok) {
            transactionInfoGrid = (
                <TransactionMetadataContext.Provider value={transactionMetadata.result}>
                    <ChainConfigContext.Provider value={chainConfig}>
                        <LabelMetadataContext.Provider value={labelMetadata}>
                            <PriceMetadataContext.Provider value={priceMetadata}>
                                <TransactionInfo estimator={estimator!} provider={provider!} />
                            </PriceMetadataContext.Provider>
                        </LabelMetadataContext.Provider>
                    </ChainConfigContext.Provider>
                </TransactionMetadataContext.Provider>
            );
        } else {
            transactionInfoGrid = <>Failed to fetch transaction: {transactionMetadata.error.toString()}</>;
        }
    }

    let valueChanges;
    if (transactionMetadata && traceResult && traceMetadata && provider) {
        if (transactionMetadata.ok) {
            valueChanges = (
                <TransactionMetadataContext.Provider value={transactionMetadata.result}>
                    <ChainConfigContext.Provider value={chainConfig}>
                        <LabelMetadataContext.Provider value={labelMetadata}>
                            <PriceMetadataContext.Provider value={priceMetadata}>
                                <TokenMetadataContext.Provider value={tokenMetadata}>
                                    <ValueChange
                                        traceResult={traceResult}
                                        traceMetadata={traceMetadata}
                                        provider={provider}
                                    />
                                </TokenMetadataContext.Provider>
                            </PriceMetadataContext.Provider>
                        </LabelMetadataContext.Provider>
                    </ChainConfigContext.Provider>
                </TransactionMetadataContext.Provider>
            );
        } else {
            transactionInfoGrid = <>Failed to fetch transaction or trace</>;
        }
    }

    let transactionActions;
    if (transactionMetadata && traceResult && traceMetadata && provider) {
        if (transactionMetadata.ok) {
            transactionActions = (
                <TransactionMetadataContext.Provider value={transactionMetadata.result}>
                    <ChainConfigContext.Provider value={chainConfig}>
                        <LabelMetadataContext.Provider value={labelMetadata}>
                            <PriceMetadataContext.Provider value={priceMetadata}>
                                <TokenMetadataContext.Provider value={tokenMetadata}>
                                    <DecodeTree
                                        traceResult={traceResult}
                                        traceMetadata={traceMetadata}
                                        provider={provider}
                                    />
                                </TokenMetadataContext.Provider>
                            </PriceMetadataContext.Provider>
                        </LabelMetadataContext.Provider>
                    </ChainConfigContext.Provider>
                </TransactionMetadataContext.Provider>
            );
        } else {
            transactionActions = <>Failed to fetch transaction</>;
        }
    }

    let traceTree;
    if (traceResult && traceMetadata) {
        traceTree = (
            <ChainConfigContext.Provider value={chainConfig}>
                <LabelMetadataContext.Provider value={labelMetadata}>
                    <PreimageMetadataContext.Provider value={preimageMetadata}>
                        <TraceTree traceResult={traceResult} traceMetadata={traceMetadata} />
                    </PreimageMetadataContext.Provider>
                </LabelMetadataContext.Provider>
            </ChainConfigContext.Provider>
        );
    }

    let content;
    if (!transactionMetadata) {
        content = (
            <>
                <Typography variant={'h6'}>Loading transaction...</Typography>
            </>
        );
    } else {
        content = (
            <>
                {transactionInfoGrid ? (
                    <>
                        <Typography variant={'h6'}>Transaction Info</Typography>
                        {transactionInfoGrid}
                    </>
                ) : null}

                {valueChanges ? (
                    <>
                        <Typography variant={'h6'}>Value Changes</Typography>
                        {valueChanges}
                    </>
                ) : null}

                {transactionActions ? (
                    <>
                        <Typography variant={'h6'}>Decoded Actions</Typography>
                        <Box
                            sx={{
                                overflowX: {
                                    xs: 'auto',
                                    md: 'inherit',
                                },
                                paddingBottom: {
                                    xs: '1em',
                                    md: 'inherit',
                                },
                            }}
                        >
                            {transactionActions}
                        </Box>
                    </>
                ) : null}

                {traceTree ? (
                    <>
                        <Typography variant={'h6'}>Call Trace</Typography>
                        <Box
                            sx={{
                                overflowX: {
                                    xs: 'auto',
                                    md: 'inherit',
                                },
                                paddingBottom: {
                                    xs: '1em',
                                    md: 'inherit',
                                },
                            }}
                        >
                            {traceTree}
                        </Box>
                    </>
                ) : null}
            </>
        );
    }

    return (
        <>
            <Navbar
                title={'Transaction Tracer'}
                description={'View and trace EVM transactions'}
                icon={'/tracer.png'}
                url={'/trace'}
                content={null}
            />

            <Box sx={{}} p={2}>
                {content}
            </Box>
        </>
    );
}
