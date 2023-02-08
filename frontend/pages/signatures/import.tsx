import CloseIcon from '@mui/icons-material/Close';
import { AlertColor, TextField } from '@mui/material';

import Alert from '@mui/material/Alert';

import Box from '@mui/material/Box';

import Button from '@mui/material/Button';
import ButtonGroup from '@mui/material/ButtonGroup';
import Collapse from '@mui/material/Collapse';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import Grid2 from '@mui/material/Unstable_Grid2';
import { ethers } from 'ethers';
import { useRouter } from 'next/router';
import React from 'react';
import Navbar from '../../components/Navbar';
import examples from '../../components/signatures/examples';

import { apiEndpoint } from '../../components/signatures/helpers';

type ImportRequest = {
    function: string[];
    event: string[];
};

type ImportResponseDetails = {
    imported: Record<string, string>;
    duplicated: Record<string, string>;
    invalid: string[];
};

type ImportResponse = {
    function: ImportResponseDetails;
    event: ImportResponseDetails;
};

function constructRequest(input: string): ImportRequest {
    try {
        let abi = new ethers.utils.Interface(input);
        return {
            function: Object.values(abi.functions).map((frag) => frag.format()),
            event: Object.values(abi.events).map((frag) => frag.format()),
        };
    } catch (e) {
        console.log(e);
    }

    return input
        .split('\n')
        .map((line) => {
            if (!line.startsWith('#define ')) return line;

            line = line.substring('#define '.length);
            let balance = 0;
            for (let i = line.indexOf('('); i < line.length; i++) {
                if (line.charAt(i) === '(') balance++;
                else if (line.charAt(i) === ')') balance--;

                if (balance === 0) {
                    line = line.substring(0, i + 1);
                    break;
                }
            }

            return line;
        })
        .reduce(
            (obj, line) => {
                if (line.startsWith('function ')) {
                    obj.function.push(line.substring('function '.length));
                } else if (line.startsWith('error ')) {
                    obj.function.push(line.substring('error '.length));
                } else if (line.startsWith('event ')) {
                    obj.event.push(line.substring('event '.length));
                }
                return obj;
            },
            { function: [] as string[], event: [] as string[] },
        );
}

function generateImportRow(sig: string, hash: string, color: string): React.ReactElement {
    return (
        <TableRow
            key={sig}
            sx={{
                '&:last-child td, &:last-child th': { border: 0 },
                backgroundColor: color,
            }}
        >
            <TableCell component="th" scope="row" align="center">
                <code>{sig}</code>
            </TableCell>
            <TableCell align="center">
                <code>{hash}</code>
            </TableCell>
        </TableRow>
    );
}

function processImport(results: ImportResponse): [React.ReactElement, string] {
    let importTableRows = [] as React.ReactElement[];

    for (let [sigType, sigResultsMap] of Object.entries(results)) {
        for (let [sig, hash] of Object.entries(sigResultsMap['imported'])) {
            importTableRows.push(generateImportRow(sig, hash, '#90ee90'));
        }
        for (let [sig, hash] of Object.entries(sigResultsMap['duplicated'])) {
            importTableRows.push(generateImportRow(sig, hash, '#d3d3d3'));
        }
        if (sigResultsMap['invalid']) {
            for (let sig of sigResultsMap['invalid']) {
                importTableRows.push(generateImportRow(sig, '', '#ff9999'));
            }
        }
    }

    let renderedImportResults = (
        <Grid item xs={12} md={10} lg={8}>
            <Box display="flex" justifyContent="center">
                <TableContainer component={Paper}>
                    <Table size="small">
                        <TableHead>
                            <TableRow>
                                <TableCell align="center">Signature</TableCell>
                                <TableCell align="center">Hash</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>{importTableRows}</TableBody>
                    </Table>
                </TableContainer>
            </Box>
        </Grid>
    );

    let impFunctions = Object.entries(results['function']['imported']).length;
    let impEvents = Object.entries(results['event']['imported']).length;
    let dupFunctions = Object.entries(results['function']['duplicated']).length;
    let dupEvents = Object.entries(results['event']['duplicated']).length;

    return [
        renderedImportResults,
        `Imported ${impFunctions} functions and ${impEvents} events! Skipped ${dupFunctions} functions and ${dupEvents} events.`,
    ];
}

export default function Import() {
    const [importData, setImportData] = React.useState('');
    const [isImporting, setIsImporting] = React.useState(false);
    const [importResults, setImportResults] = React.useState(<div></div>);
    const [alertData, setAlertData] = React.useState({
        dismissed: true,
        severity: 'success' as AlertColor,
        message: '',
    });

    const router = useRouter();

    const submitImport = () => {
        // check if the imported data is empty
        const importDataTrimmed = importData.trim();
        if (importDataTrimmed.length === 0) {
            setAlertData({
                dismissed: false,
                severity: 'warning',
                message: `The import data field is empty, please provide some data.`,
            });
            return;
        }

        setIsImporting(true);
        setAlertData((prevState) => ({
            ...prevState,
            dismissed: true,
        }));
        fetch(`${apiEndpoint()}/v1/import`, {
            method: 'POST',
            body: JSON.stringify(constructRequest(importDataTrimmed)),
        })
            .then((res) => res.json())
            .then((json) => {
                if (json['ok'] === false) {
                    throw new Error(json['error']);
                }

                let [importTable, alertMessage] = processImport(json['result'] as ImportResponse);

                setIsImporting(false);
                setImportResults(importTable);
                setAlertData({
                    dismissed: false,
                    severity: 'success',
                    message: alertMessage,
                });
            })
            .catch((e) => {
                setIsImporting(false);
                setAlertData({
                    dismissed: false,
                    severity: 'error',
                    message: `An error occurred: ${e.message}`,
                });
            });
    };

    return (
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
                <Grid2 container p={2} gap={2}>
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
                        <Box display="flex" justifyContent="center">
                            <TextField
                                id="import"
                                type="text"
                                placeholder="Enter data to import or view an example"
                                sx={{
                                    width: {
                                        xs: '90vw',
                                        md: '75vw',
                                        lg: '60vw',
                                    },
                                }}
                                inputProps={{
                                    style: {
                                        fontFamily: 'monospace',
                                    },
                                }}
                                minRows={16}
                                multiline={true}
                                value={importData}
                                onChange={(event) => setImportData(event.target.value)}
                            />
                        </Box>
                    </Grid2>

                    <Grid2 xs={12}>
                        <Grid2 container spacing={2} alignItems="center" justifyContent="center">
                            <Grid2 xs={'auto'}>
                                <Box display="flex" justifyContent="center">
                                    <Typography variant="button">Examples</Typography>
                                </Box>
                            </Grid2>
                            <Grid2 xs={'auto'}>
                                <ButtonGroup variant="outlined">
                                    <Button onClick={() => setImportData(examples.raw)}>Raw</Button>
                                    <Button onClick={() => setImportData(examples.abi)}>ABI</Button>
                                    <Button onClick={() => setImportData(examples.huff)}>Huff</Button>
                                </ButtonGroup>
                            </Grid2>
                        </Grid2>
                    </Grid2>

                    <Grid2 xs={12}>
                        <Box display="flex" justifyContent="center">
                            <Button variant="contained" onClick={submitImport} disabled={isImporting}>
                                Submit
                            </Button>
                        </Box>
                    </Grid2>

                    {importResults}
                </Grid2>
            }
        />
    );
}
