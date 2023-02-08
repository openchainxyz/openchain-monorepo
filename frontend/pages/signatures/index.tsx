import { useRouter } from 'next/router';
import CloseIcon from '@mui/icons-material/Close';

import Box from '@mui/material/Box';
import { Alert, AlertColor, Collapse, TextField } from '@mui/material';
import React, { useState } from 'react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';

import { apiEndpoint } from '../../components/signatures/helpers';
import Navbar from '../../components/Navbar';
import Grid2 from '@mui/material/Unstable_Grid2';

type SignatureData = {
    name: string;
    filtered: boolean;
};

type SignaturesResponse = {
    function: Record<string, SignatureData[]>;
    event: Record<string, SignatureData[]>;
};

function constructSearchParams(query: string): [string, string, 'lookup_function' | 'lookup_event' | 'search'] {
    const hexRe = /^[0-9A-Fa-f]*$/;

    let params = new URLSearchParams();
    params.append('filter', 'false');

    if (query.length === 10 && hexRe.test(query.substring(2))) {
        params.append('function', query);
        return ['lookup', params.toString(), 'lookup_function'];
    } else if (query.length === 66 && hexRe.test(query.substring(2))) {
        params.append('event', query);
        return ['lookup', params.toString(), 'lookup_event'];
    }

    if (query.indexOf('*') === -1 && query.indexOf('?') === -1) {
        query = query + '*';
    }
    params.append('query', query);
    return ['search', params.toString(), 'search'];
}

function processSearch(
    method: 'lookup_function' | 'lookup_event' | 'search',
    results: SignaturesResponse,
): React.ReactNode {
    let searchTableRows = [] as React.ReactElement[];

    for (let [sigType, sigResultsMap] of Object.entries(results)) {
        for (let [sig, sigResults] of Object.entries(sigResultsMap)) {
            if (sigResults.length === 0) {
                searchTableRows.push(
                    <TableRow
                        key={sig}
                        sx={{
                            '&:last-child td, &:last-child th': { border: 0 },
                            backgroundColor: '#ff9999',
                        }}
                    >
                        <TableCell component="th" scope="row" align="center">
                            <code>{sig}</code>
                        </TableCell>
                        <TableCell align="center"></TableCell>
                    </TableRow>,
                );
            } else {
                for (let sigResult of sigResults) {
                    let isFiltered = sigResult['filtered'] || sigResult['filtered'];

                    searchTableRows.push(
                        <TableRow
                            key={sig + sigType + sigResult['name']}
                            sx={{
                                '&:last-child td, &:last-child th': { border: 0 },
                                backgroundColor: (theme) => (isFiltered ? '#d3d3d3' : '#90ee90'),
                            }}
                        >
                            <TableCell component="th" scope="row">
                                <IconButton
                                    aria-label="copy"
                                    color="inherit"
                                    size="small"
                                    onClick={() => {
                                        navigator.clipboard.writeText(sig);
                                    }}
                                >
                                    <ContentCopyIcon fontSize={'small'}></ContentCopyIcon>
                                </IconButton>
                                <code>{sig}</code>
                            </TableCell>
                            <TableCell>
                                <IconButton
                                    aria-label="copy"
                                    color="inherit"
                                    size="small"
                                    onClick={() => {
                                        navigator.clipboard.writeText(sigResult['name']);
                                    }}
                                >
                                    <ContentCopyIcon fontSize={'small'}></ContentCopyIcon>
                                </IconButton>
                                <code>{sigResult['name']}</code>
                            </TableCell>
                        </TableRow>,
                    );
                }
            }
        }
    }

    return (
        <Grid2>
            <Box display="flex" justifyContent="center">
                <TableContainer component={Paper}>
                    <Table size="small">
                        <TableHead>
                            <TableRow>
                                <TableCell>Hash</TableCell>
                                <TableCell>Name</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>{searchTableRows}</TableBody>
                    </Table>
                </TableContainer>
            </Box>
        </Grid2>
    );
}

export default function Index() {
    const [query, setQuery] = useState('');
    const [isSearching, setIsSearching] = useState(false);
    const [searchResults, setSearchResults] = useState<React.ReactNode>(<div></div>);
    const [alertData, setAlertData] = React.useState({
        dismissed: true,
        severity: 'success' as AlertColor,
        message: '',
    });

    const doSearch = () => {
        console.log('fuck');
        // check if the imported data is empty
        const queryTrimmed = query.trim();
        if (queryTrimmed.length === 0) {
            console.log('fuck2');
            setAlertData({
                dismissed: false,
                severity: 'warning',
                message: `The search field is empty.`,
            });
            return;
        }

        setIsSearching(true);
        setAlertData((prevState) => ({
            ...prevState,
            dismissed: true,
        }));
        const [method, params, searchType] = constructSearchParams(queryTrimmed);
        fetch(`${apiEndpoint()}/v1/${method}?${params}`)
            .then((res) => res.json())
            .then((json) => {
                if (json['ok'] === false) {
                    throw new Error(json['error']);
                }

                setIsSearching(false);
                setSearchResults(processSearch(searchType, json['result'] as SignaturesResponse));
            })
            .catch((e) => {
                setIsSearching(false);
                setAlertData({
                    dismissed: false,
                    severity: 'error',
                    message: `An error occurred: ${e.message}`,
                });
            });
    };

    return (
        <>
            <Navbar
                title={'Signature Database'}
                description={'Search for or import function signatures'}
                icon={'/signatures.png'}
                url={'/signatures'}
                links={[
                    {
                        name: 'Import',
                        url: '/signatures/import',
                    },
                ]}
                content={
                    <>
                        <Grid2 container p={2}>
                            <Grid2 xs={12}>
                                <Collapse in={!alertData.dismissed}>
                                    <Box display="flex" justifyContent="center">
                                        <Alert
                                            severity={alertData.severity}
                                            action={
                                                <IconButton
                                                    aria-label="close"
                                                    color="inherit"
                                                    size="small"
                                                    onClick={() => {
                                                        setAlertData((prevState) => ({
                                                            ...prevState,
                                                            dismissed: true,
                                                        }));
                                                    }}
                                                >
                                                    <CloseIcon fontSize="inherit" />
                                                </IconButton>
                                            }
                                            sx={{ mb: 2 }}
                                        >
                                            {alertData.message}
                                        </Alert>
                                    </Box>
                                </Collapse>
                            </Grid2>
                            <Grid2 xs={12}>
                                <TextField
                                    variant="standard"
                                    placeholder="Enter selectors or text..."
                                    fullWidth
                                    margin="dense"
                                    value={query}
                                    onChange={(event) => setQuery(event.target.value)}
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
                                                Search
                                            </Button>
                                        ),
                                    }}
                                ></TextField>
                            </Grid2>
                        </Grid2>
                    </>
                }
            />

            <Box sx={{}} p={2}>
                {searchResults}
            </Box>
        </>
    );
}
