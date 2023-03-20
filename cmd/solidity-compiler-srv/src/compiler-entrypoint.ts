import * as child_process from 'child_process';
import { Readable } from 'stream';
import { StandardInput } from './types';
import { translateJsonCompilerOutput } from 'solc/translate';

import setupMethods, { Solc } from 'solc/wrapper';

const transformLinkReferences = (linkReferences: any): any => {
    if (linkReferences === undefined) {
        return undefined;
    }

    const newLinkReferences: Record<string, any> = {};
    const topLevelLinkReferences: Record<string, any> = {};

    for (const [k, v] of Object.entries(linkReferences)) {
        if (Array.isArray(v)) {
            topLevelLinkReferences[k] = v;
        } else {
            newLinkReferences[k] = v;
        }
    }

    if (Object.entries(topLevelLinkReferences).length === 0) {
        return newLinkReferences;
    }

    newLinkReferences[''] = topLevelLinkReferences;

    return newLinkReferences;
};

const transformLegacyResponse = (output: string, isOptimized: boolean, compiledSeeds: any[]): any => {
    const transformedResponse = translateJsonCompilerOutput(JSON.parse(output));

    const contracts = transformedResponse.contracts;
    if (contracts) {
        for (const fileName of Object.keys(contracts)) {
            for (const contractName of Object.keys(contracts[fileName])) {
                const contractData = contracts[fileName][contractName];

                contractData['metadata'] = JSON.stringify({
                    settings: {
                        optimizer: {
                            enabled: isOptimized,
                        },
                        nonDeterministicSeeds: compiledSeeds.map((seed) => {
                            return {
                                sources: seed['sources'],
                                settings: seed['settings'],
                            };
                        }),
                    },
                });

                const evm = contractData['evm'];

                const bytecode = evm['bytecode'];
                if (bytecode !== undefined) {
                    bytecode['linkReferences'] = transformLinkReferences(bytecode['linkReferences']);
                }

                const deployedBytecode = evm['deployedBytecode'];
                if (deployedBytecode !== undefined) {
                    deployedBytecode['linkReferences'] = transformLinkReferences(deployedBytecode['linkReferences']);
                }
            }
        }
    }

    return transformedResponse;
};

const compileInProcess = (solc: Solc, input: StandardInput): Promise<any> => {
    if (solc.lowlevel.compileStandard !== null) {
        return JSON.parse(solc.lowlevel.compileStandard(JSON.stringify(input)));
    }

    const legacySources = Object.entries(input.sources).reduce((sources, [path, source]) => {
        if (source.content === undefined || source.content === null) {
            throw new Error('missing content for source: ' + path);
        }
        sources[path] = source.content;
        return sources;
    }, {} as Record<string, string>);
    const isOptimized = input.settings?.optimizer?.enabled;
    const nonDeterministicSeeds = input.settings?.nonDeterministicSeeds;

    const compiledSeeds = [];
    if (Array.isArray(nonDeterministicSeeds)) {
        for (const seed of nonDeterministicSeeds) {
            compiledSeeds.push(compileInProcess(solc, seed));
        }
    }

    if (solc.lowlevel.compileCallback !== null) {
        const compileCallbackResponse = solc.lowlevel.compileCallback(
            JSON.stringify({ sources: legacySources }),
            isOptimized,
        );
        return transformLegacyResponse(compileCallbackResponse, isOptimized, compiledSeeds);
    }

    if (solc.lowlevel.compileMulti !== null) {
        const compileMultiResponse = solc.lowlevel.compileMulti(
            JSON.stringify({ sources: legacySources }),
            isOptimized,
        );
        return transformLegacyResponse(compileMultiResponse, isOptimized, compiledSeeds);
    }

    if (solc.lowlevel.compileSingle !== null) {
        const legacySourceEntries = Object.entries(legacySources);
        if (legacySourceEntries.length > 1) {
            throw new Error('multiple sources not supported in legacy mode');
        }

        const compileSingleResponse = solc.lowlevel.compileSingle(legacySourceEntries[0][1], isOptimized);
        return transformLegacyResponse(compileSingleResponse, isOptimized, compiledSeeds);
    }

    throw new Error('compiler does not support any json interfaces');
};

const readAll = async (stream: Readable): Promise<string> => {
    const chunks = [];
    for await (const chunk of stream) {
        chunks.push(chunk);
    }
    return Buffer.concat(chunks).toString('utf-8');
};

readAll(process.stdin).then((contents) => {
    const solc = setupMethods(require(process.argv[2]));
    const output = compileInProcess(solc, JSON.parse(contents));
    process.stdout.write(JSON.stringify(output));
});
