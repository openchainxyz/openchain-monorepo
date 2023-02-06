import { useRouter } from 'next/router';
import { MenuItem, TextField } from '@mui/material';
import React from 'react';
import Button from '@mui/material/Button';
import Navbar from '../../components/Navbar';
import Grid2 from '@mui/material/Unstable_Grid2';
import { SupportedChains } from '../../components/tracer/Chains';

export default function Index() {
    const router = useRouter();
    const { chain: queryChain, txhash: queryTxhash } = router.query;

    // sets the default chain to ethereum.
    const [chain, setChain] = React.useState('ethereum');
    const [txhash, setTxhash] = React.useState('');

    React.useEffect(() => {
        if (!queryChain || Array.isArray(queryChain)) return;
        if (!queryTxhash || Array.isArray(queryTxhash)) return;

        setChain(queryChain);
        setTxhash(queryTxhash);
    }, [queryChain, queryTxhash]);

    const doSearch = () => {
        if (/0x[0-9a-fA-F]{64}/g.test(txhash)) {
            router.push(`/trace/${chain}/${txhash}`);
        }
    };

    return (
        <Navbar
            title={'Transaction Tracer'}
            description={'View and trace EVM transactions'}
            icon={'/tracer.png'}
            url={'/trace'}
            content={
                <>
                    <Grid2 container p={2}>
                        <Grid2>
                            <TextField
                                onChange={(event) => setChain(event.target.value)}
                                value={chain}
                                variant="standard"
                                select
                                margin="dense"
                                fullWidth
                                SelectProps={{
                                    style: {
                                        fontFamily: 'RiformaLL',
                                    },
                                }}
                            >
                                {SupportedChains.map((v) => {
                                    return (
                                        <MenuItem key={v.id} value={v.id} style={{ fontFamily: 'RiformaLL' }}>
                                            {v.displayName}
                                        </MenuItem>
                                    );
                                })}
                                {!SupportedChains.find((sChain) => sChain.id === chain) ? (
                                    <MenuItem key={chain} value={chain} style={{ fontFamily: 'RiformaLL' }}>
                                        {queryChain}
                                    </MenuItem>
                                ) : null}
                            </TextField>
                        </Grid2>
                        <Grid2 xs>
                            <TextField
                                variant="standard"
                                placeholder="Enter txhash..."
                                fullWidth
                                margin="dense"
                                onChange={(event) => setTxhash(event.target.value)}
                                value={txhash}
                                onKeyUp={(event) => {
                                    if (event.key === 'Enter') {
                                        doSearch();
                                    }
                                }}
                                inputProps={{
                                    style: {
                                        fontFamily: 'RiformaLL',
                                    },
                                }}
                                InputProps={{
                                    endAdornment: (
                                        <Button
                                            variant="text"
                                            size="small"
                                            onClick={() => doSearch()}
                                            style={{
                                                fontFamily: 'RiformaLL',
                                            }}
                                        >
                                            View
                                        </Button>
                                    ),
                                }}
                            ></TextField>
                        </Grid2>
                    </Grid2>
                </>
            }
        />
    );
}
