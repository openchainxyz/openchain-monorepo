import cors from 'cors';
import express from 'express';

import fs from 'fs';

import { CompileRequest } from './types';

import * as child_process from 'child_process';

const allVersions = new Set<String>();

const downloadSolc = async (version: string) => {
    const filename = `/tmp/soljson-${version}.js`;

    try {
        await fs.promises.stat(filename);
    } catch {
        const fetchResult = await fetch(`https://binaries.soliditylang.org/bin/soljson-${version}.js`);
        if (fetchResult.status !== 200) {
            throw new Error(`fetch ${version} returned ${fetchResult.status}: ${fetchResult.statusText}`);
        }

        let src = await fetchResult.text();

        if (src.includes(`"use asm";`)) {
            // asm.js doesn't work on node >14 (seems like v8 is a lot stricter or it never worked in the first place)
            // just remove it to avoid confusing warnings
            src = src.replace(`"use asm";`, '');
        }

        await fs.promises.writeFile(filename, src);
    }
};

const refreshSolidityVersions = async () => {
    const result = await fetch('https://raw.githubusercontent.com/ethereum/solc-bin/gh-pages/bin/list.txt');
    const text = await result.text();

    text.split('\n')
        .map((line) => {
            return line.replace('soljson-', '').replace('.js', '');
        })
        .forEach((v) => allVersions.add(v));

    setTimeout(refreshSolidityVersions, 86400);
};

refreshSolidityVersions();

const app = express();
app.use(cors());
app.use(express.json({ limit: 1e9 }));

app.post('/v1/compile', async (req: any, res: any) => {
    const body = req.body as CompileRequest;

    if (typeof body.version !== 'string') {
        res.send({
            ok: false,
            error: 'version is not a string',
        });
        return;
    }

    if (typeof body.input !== 'object') {
        res.send({
            ok: false,
            error: 'input is not an object',
        });
        return;
    }

    if (!allVersions.has(body.version)) {
        res.send({
            ok: false,
            error: 'invalid version',
        });
        return;
    }

    console.log(`serving compile version=${body.version}`);

    await downloadSolc(body.version);

    const child = child_process.fork('./compiler-entrypoint', [`/tmp/soljson-${body.version}.js`], {
        stdio: 'pipe',
    });
    child.stdin!.write(JSON.stringify(body.input));
    child.stdin!.end();

    const chunks = [];
    for await (const chunk of child.stdout!) {
        chunks.push(chunk);
    }

    res.send({
        ok: true,
        result: JSON.parse(Buffer.concat(chunks).toString('utf8')),
    });
});

app.listen(3000, () => {
    console.log('listening on port 3000');
});
