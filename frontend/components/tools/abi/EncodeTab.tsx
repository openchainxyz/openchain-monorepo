import { TextField } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2';
import { AbiCoder, FunctionFragment } from 'ethers';
import { useRouter } from 'next/router';
import React, { useEffect, useState } from 'react';

export function EncodeTab() {
    const [data, setData] = useState<string>('');
    const [paramTypes, setParamTypes] = useState<string>('');
    const [encodeResult, setEncodeResult] = useState<string>('');

    const router = useRouter();

    const asPath = router.asPath;

    useEffect(() => {
        const hash = asPath.split('#')[1];
    }, [asPath]);

    useEffect(() => {
        if (!data.length || !paramTypes.length) {
            setEncodeResult('');
            return;
        }

        let parsedData: any;
        try {
            parsedData = JSON.parse(data);
        } catch (e: any) {
            setEncodeResult('Failed to decode data: ' + e.toString());
            return;
        }

        let fragment: FunctionFragment;
        try {
            fragment = FunctionFragment.from(paramTypes);
        } catch (e: any) {
            setEncodeResult('Failed to parse parameters: ' + e.toString());
            return;
        }

        try {
            setEncodeResult(
                fragment.selector + AbiCoder.defaultAbiCoder().encode(fragment.inputs, parsedData).substring(2),
            );
        } catch (e: any) {
            setEncodeResult('Failed to encode data: ' + e.toString());
        }
    }, [data, paramTypes]);

    return (
        <Grid2 container gap={2}>
            <Grid2 xs={12}>
                <TextField
                    variant="standard"
                    label="Raw Data (JSON)"
                    fullWidth
                    margin="dense"
                    value={data}
                    multiline
                    onChange={(event) => setData(event.target.value)}
                    inputProps={{
                        style: {
                            fontFamily: 'RiformaLL',
                        },
                    }}
                ></TextField>
            </Grid2>
            <Grid2 xs={12}>
                <TextField
                    variant="standard"
                    label="Parameters"
                    placeholder={'func(uint,uint,string)'}
                    fullWidth
                    margin="dense"
                    value={paramTypes}
                    multiline
                    onChange={(event) => setParamTypes(event.target.value)}
                    inputProps={{
                        style: {
                            fontFamily: 'RiformaLL',
                        },
                    }}
                ></TextField>
            </Grid2>
            <Grid2 xs={12}>
                <TextField
                    variant="standard"
                    label={'Results'}
                    fullWidth
                    margin="dense"
                    value={encodeResult}
                    multiline
                    inputProps={{
                        style: {
                            fontFamily: 'RiformaLL',
                        },
                    }}
                ></TextField>
            </Grid2>
        </Grid2>
    );
}
