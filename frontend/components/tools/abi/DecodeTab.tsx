import { ParamTreeView } from '@components/tracer/ParamTreeView';
import { Checkbox, FormControl, FormControlLabel, Radio, RadioGroup, TextField } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2';
import { guessAbiEncodedData, guessFragment } from '@openchainxyz/abi-guesser';
import { AbiCoder, FunctionFragment, Interface } from 'ethers';
import { useRouter } from 'next/router';
import React, { useEffect, useState } from 'react';

export function DecodeTab() {
    const [data, setData] = useState<string>('');
    const [paramTypes, setParamTypes] = useState<string>('');
    const [decodeMode, setDecodeMode] = useState<'manual' | 'auto'>('auto');
    const [isFunction, setIsFunction] = useState<boolean>(true);
    const [treeView, setTreeView] = useState<React.ReactNode>(null);

    const [decodedResult, setDecodedResult] = useState<string>('');

    const router = useRouter();

    const asPath = router.asPath;

    useEffect(() => {
        const hash = asPath.split('#')[1];
    }, [asPath]);

    const tryGuessData = (data: string, isFunction: boolean) => {
        const strippedData = data.replaceAll('\n', '');

        try {
            if (!isFunction) {
                const paramTypes = guessAbiEncodedData(strippedData);
                if (paramTypes) {
                    setParamTypes(`func(${paramTypes.map((v) => v.format()).join(',')})`);
                } else {
                    setParamTypes('');
                }
            } else {
                const fragment = guessFragment(strippedData);
                if (fragment) {
                    setParamTypes(fragment.format());
                } else {
                    setParamTypes('');
                }
            }
        } catch (e) {
            setParamTypes('');
        }
    };

    const onDataChanged = (newData: string) => {
        setData(newData);
    };

    useEffect(() => {
        if (!data.length) {
            setParamTypes('');
            return;
        }

        if (decodeMode === 'auto') {
            if (isFunction) {
                fetch(
                    `https://api.openchain.xyz/signature-database/v1/lookup?function=${data.substring(
                        0,
                        10,
                    )}&filter=true`,
                )
                    .then((res) => res.json())
                    .then((json) => {
                        if (json['ok'] === false) {
                            throw new Error(json['error']);
                        }

                        const results = json['result']['function'][data.substring(0, 10)];

                        if (!results || !results.length) {
                            tryGuessData(data, true);
                            return;
                        }

                        setParamTypes(results[0].name);
                    })
                    .catch((e) => {
                        console.log('failed to search', e);
                        tryGuessData(data, true);
                    });
            } else {
                tryGuessData(data, false);
            }
        }
    }, [data, decodeMode, isFunction]);

    useEffect(() => {
        if (!data.length || !paramTypes.length) {
            setDecodedResult('');
            setTreeView(null);
            return;
        }

        let strippedData = data.replaceAll('\n', '');
        if (strippedData.startsWith('0x')) strippedData = strippedData.substring(2);

        let fragment: FunctionFragment | undefined;

        let abi: any[] | undefined;
        try {
            abi = JSON.parse(paramTypes);
        } catch {}
        if (abi) {
            const intf = new Interface(abi);
            const functions = intf.fragments.filter((frag) => frag instanceof FunctionFragment);
            if (functions.length === 1) {
                fragment = functions[0] as FunctionFragment;
            } else if (isFunction) {
                try {
                    const intfFragment = intf.getFunction(strippedData.substring(0, 8));
                    if (!intfFragment) throw new Error('not found');

                    fragment = intfFragment;
                } catch {
                    setDecodedResult('Could not find function fragment');
                    setTreeView(null);
                    return;
                }
            }
        }

        if (!fragment) {
            try {
                fragment = FunctionFragment.from(paramTypes);
            } catch (e: any) {
                setDecodedResult('Failed to parse parameters: ' + e.toString());
                setTreeView(null);
                return;
            }
        }

        if (isFunction) {
            strippedData = strippedData.substring(8);
        }

        try {
            const values = AbiCoder.defaultAbiCoder().decode(fragment.inputs, '0x' + strippedData);
            setDecodedResult(
                JSON.stringify(values, (key, value) => (typeof value === 'bigint' ? value.toString() : value), 2),
            );

            setTreeView(<ParamTreeView path={''} params={fragment.inputs} values={values} />);
        } catch (e: any) {
            setDecodedResult('Failed to decode data: ' + e.toString());
            setTreeView(null);
        }
    }, [data, paramTypes, isFunction]);

    return (
        <Grid2 container gap={2}>
            <Grid2 xs={12}>
                <TextField
                    variant="standard"
                    label="ABI Data"
                    fullWidth
                    margin="dense"
                    value={data}
                    multiline
                    maxRows={8}
                    onChange={(event) => onDataChanged(event.target.value)}
                    inputProps={{
                        style: {
                            fontFamily: 'RiformaLL',
                        },
                    }}
                ></TextField>
            </Grid2>
            <Grid2 xs={12}>
                <FormControl>
                    <FormControlLabel
                        control={
                            <Checkbox checked={isFunction} onChange={(event) => setIsFunction(event.target.checked)} />
                        }
                        label="Assume first four bytes are function selector"
                    />
                    <RadioGroup
                        value={decodeMode}
                        onChange={(event) => {
                            setDecodeMode(event.target.value as 'manual' | 'auto');
                        }}
                    >
                        <FormControlLabel
                            value="auto"
                            control={<Radio />}
                            label="Automatically determine the best parameter types"
                        />
                        <FormControlLabel value="manual" control={<Radio />} label="Manually specify parameter types" />
                    </RadioGroup>
                </FormControl>
            </Grid2>
            <Grid2 xs={12}>
                <TextField
                    variant="standard"
                    label="Parameters"
                    placeholder={decodeMode === 'manual' ? 'Enter function signature or ABI' : ''}
                    fullWidth
                    margin="dense"
                    maxRows={8}
                    value={paramTypes}
                    disabled={decodeMode === 'auto'}
                    multiline
                    onChange={(event) => setParamTypes(event.target.value)}
                    inputProps={{
                        style: {
                            fontFamily: 'RiformaLL',
                        },
                    }}
                ></TextField>
            </Grid2>
            <Grid2
                xs={12}
                sx={{
                    overflowX: 'scroll',
                }}
            >
                {treeView}
            </Grid2>
            <Grid2 xs={12}>
                <TextField
                    variant="standard"
                    label={'Results'}
                    fullWidth
                    margin="dense"
                    value={decodedResult}
                    multiline
                    inputProps={{
                        style: {
                            fontFamily: 'RiformaLL',
                            whiteSpace: 'pre',
                            overflowX: 'scroll',
                        },
                    }}
                ></TextField>
            </Grid2>
        </Grid2>
    );
}
