import cors from 'cors';
import express from 'express';

import fs from 'fs';
import * as path from 'path';
import { pipeline, Readable } from 'stream';

import { CompileRequest } from './types';

import * as child_process from 'child_process';
import semver = require('semver/preload');

const allVersions = new Set<String>();
const nativeVersions = new Set<String>();

const port = process.env.HTTP_PORT || '3000';
const downloadDir = process.env.DOWNLOAD_DIR || '/tmp';

const urlForVersion = (version: string): string => {
    if (nativeVersions.has(version)) {
        return `https://binaries.soliditylang.org/linux-amd64/solc-linux-amd64-${version}`;
    }

    return `https://binaries.soliditylang.org/bin/soljson-${version}.js`;
};

const downloadSolc = async (version: string): Promise<string> => {
    const downloadUrl = urlForVersion(version);

    const downloadPath = path.join(downloadDir, path.basename(downloadUrl));

    try {
        await fs.promises.stat(downloadPath);
    } catch {
        const fetchResult = await fetch(downloadUrl);
        if (fetchResult.status !== 200) {
            throw new Error(`fetch ${version} returned ${fetchResult.status}: ${fetchResult.statusText}`);
        }
        let body = await fetchResult.arrayBuffer();

        await fs.promises.writeFile(downloadPath, Buffer.from(body));

        await fs.promises.chmod(downloadPath, '755');
    }

    return downloadPath;
};

const refreshSolidityVersions = async () => {
    const newVersions = (
        await (await fetch('https://raw.githubusercontent.com/ethereum/solc-bin/gh-pages/bin/list.txt')).text()
    )
        .split('\n')
        .map((line) => line.match(/^soljson-(.*).js$/)![1]);

    allVersions.clear();
    newVersions.forEach((v) => allVersions.add(v));

    const newNativeVersions = (
        await (await fetch('https://raw.githubusercontent.com/ethereum/solc-bin/gh-pages/linux-amd64/list.txt')).text()
    )
        .split('\n')
        .map((line) => line.match(/^solc-linux-amd64-(.*)$/)![1])
        // standard json input was introduced in v0.4.11. to make it simpler for now let's just ignore anything older
        .filter((version) => semver.gte(version, 'v0.4.11'));

    nativeVersions.clear();
    newNativeVersions.forEach((v) => nativeVersions.add(v));

    setTimeout(refreshSolidityVersions, 86400);
};

refreshSolidityVersions();

const readAll = async (stream: Readable): Promise<string> => {
    const chunks = [];
    for await (const chunk of stream) {
        chunks.push(chunk);
    }
    return Buffer.concat(chunks).toString('utf-8');
};

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

    const trace = Math.random().toString(36).substring(2);

    console.log(`serving compile trace=${trace} version=${body.version}`);

    const solcPath = await downloadSolc(body.version);

    console.log(`using binary trace=${trace} path=${solcPath}`);

    const startTime = Date.now();

    let child;
    if (solcPath.endsWith('.js')) {
        child = child_process.fork('./compiler-entrypoint', [solcPath], {
            stdio: 'pipe',
        });
    } else {
        child = child_process.spawn(solcPath, ['--standard-json'], {
            stdio: 'pipe',
        });
    }

    const input = body.input;
    if (input.settings) {
        delete input.settings['modelChecker'];
    }

    child.stdin!.write(JSON.stringify(input));
    child.stdin!.end();

    const stdout = await readAll(child.stdout!);
    const stderr = await readAll(child.stderr!);

    const endTime = Date.now();

    if (stdout.length === 0) {
        console.log(`child process crashed trace=${trace} elapsed=${(endTime - startTime) / 1000}s stderr=${stderr}`);

        res.send({
            ok: false,
            error: stderr,
        });
    } else {
        console.log(`finished compiling trace=${trace} elapsed=${(endTime - startTime) / 1000}s`);

        res.send({
            ok: true,
            result: JSON.parse(stdout),
        });
    }
});

app.listen(parseInt(port), () => {
    console.log(`listening on port ${port}`);
});
