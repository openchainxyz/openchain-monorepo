import cors from 'cors';
import express from 'express';

import fs from 'fs';

import importFresh from 'import-fresh';
import semver from 'semver';
import {translateJsonCompilerOutput} from 'solc/translate';

import setupMethods, {Solc} from 'solc/wrapper';
import {CompileRequest, StandardInput} from "./types";

const allVersions = new Set<String>();

const solcs: Record<string, Solc> = {};

const getOrLoadSoljson = async (version: string): Promise<string> => {
    if (!allVersions.has(version)) throw new Error("invalid version");

    const filename = `/tmp/soljson-${version}.js`;

    try {
        fs.statSync(filename);
    } catch {
        const fetchResult = await fetch(`https://binaries.soliditylang.org/bin/soljson-${version}.js`);
        if (fetchResult.status !== 200) {
            throw new Error(`fetch ${version} returned ${fetchResult.status}: ${fetchResult.statusText}`)
        }

        let src = await fetchResult.text();

        if (src.includes(`"use asm";`)) {
            // asm.js doesn't work on node >14 (seems like v8 is a lot stricter or it never worked in the first place)
            // just remove it to avoid confusing warnings
            src = src.replace(`"use asm";`, '');
        }

        fs.writeFileSync(filename, src);
    }

    return filename;
};

const loadSolc = async (version: string): Promise<Solc> => {
    if (!solcs[version]) {
        const result = setupMethods(importFresh(await getOrLoadSoljson(version)));
        if (semver.lt(version, "v0.4.0")) {
            return result;
        }
        solcs[version] = result;
    }
    return solcs[version];
};

const transformLegacyResponse = (output: string, isOptimized: boolean, compiledSeeds: any[]): any => {
    const transformedResponse = translateJsonCompilerOutput(JSON.parse(output));

    const contracts = transformedResponse.contracts;
    if (contracts) {
        for (const fileName of Object.keys(contracts)) {
            for (const contractName of Object.keys(contracts[fileName])) {
                contracts[fileName][contractName]['metadata'] = JSON.stringify({
                    settings: {
                        optimizer: {
                            enabled: isOptimized,
                        },
                        nonDeterministicSeeds: compiledSeeds.map(seed => {
                            return {
                                sources: seed['sources'],
                                settings: seed['settings'],
                            }
                        }),
                    }
                })
            }
        }
    }

    return transformedResponse;
}

const compile = (solc: Solc, input: StandardInput): Promise<any> => {
    if (solc.lowlevel.compileStandard !== null) {
        return JSON.parse(solc.lowlevel.compileStandard(JSON.stringify(input)));
    }

    const legacySources = Object.entries(input.sources).reduce((sources, [path, source]) => {
        if (source.content === undefined || source.content === null) {
            throw new Error("missing content for source: " + path);
        }
        sources[path] = source.content;
        return sources;
    }, {} as Record<string, string>);
    const isOptimized = input.settings?.optimizer?.enabled;
    const nonDeterministicSeeds = input.settings?.nonDeterministicSeeds;

    const compiledSeeds = [];
    if (Array.isArray(nonDeterministicSeeds)) {
        for (const seed of nonDeterministicSeeds) {
            compiledSeeds.push(compile(solc, seed));
        }
    }

    if (solc.lowlevel.compileCallback !== null) {
        const compileCallbackResponse = solc.lowlevel.compileCallback(JSON.stringify({sources: legacySources}), isOptimized);
        return transformLegacyResponse(compileCallbackResponse, isOptimized, compiledSeeds);
    }

    if (solc.lowlevel.compileMulti !== null) {
        const compileMultiResponse = solc.lowlevel.compileMulti(JSON.stringify({sources: legacySources}), isOptimized);
        return transformLegacyResponse(compileMultiResponse, isOptimized, compiledSeeds);
    }

    if (solc.lowlevel.compileSingle !== null) {
        const legacySourceEntries = Object.entries(legacySources);
        if (legacySourceEntries.length > 1) {
            throw new Error("multiple sources not supported in legacy mode");
        }

        const compileSingleResponse = solc.lowlevel.compileSingle(legacySourceEntries[0][1], isOptimized);
        return transformLegacyResponse(compileSingleResponse, isOptimized, compiledSeeds);
    }

    throw new Error("compiler does not support any json interfaces");
};

const loadSolcVersions = async () => {
    const result = await fetch("https://raw.githubusercontent.com/ethereum/solc-bin/gh-pages/bin/list.txt");
    const text = await result.text();

    return text.split("\n").map(line => {
        return line.replace("soljson-", "").replace(".js", "");
    });
}

setInterval(async () => {
    const latestVersions = await loadSolcVersions();
    latestVersions.forEach(v => allVersions.add(v))
}, 86400)

const app = express();
app.use(cors());
app.use(express.json({limit: 1e9}));

app.post('/v1/compile', async (req, res) => {
    const body = req.body as CompileRequest;

    try {
        const solc = await loadSolc(body.version);
        const output = compile(solc, body.input);
        res.send(output);
    } catch (e: any) {
        console.log("error", e);
        res.status(500).send({ok: false, error: 'failed to compile: ' + e.toString()});
        return
    }
});

app.use((err: any, req: any, res: any, next: any) => {
    console.log("caught error!")
    console.error(err.stack)
    res.status(500).send('internal error')
})

app.listen(3000, () => {
    console.log("listening on port 3000")
})
