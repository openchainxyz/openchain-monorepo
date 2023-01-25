import Head from 'next/head';
import { useRouter } from 'next/router';

import ResponseAppBar from '../../components/ResponseAppBar';

import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import { AlertColor, TextField } from '@mui/material';
import React, { useState } from 'react';

import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Button from '@mui/material/Button';

import Alert from '@mui/material/Alert';
import IconButton from '@mui/material/IconButton';
import Collapse from '@mui/material/Collapse';
import CloseIcon from '@mui/icons-material/Close';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';

import { apiEndpoint } from '../../components/helpers';

type SignatureData = {
  name: string;
  filtered: boolean;
};

type SignaturesResponse = {
  function: Record<string, SignatureData[]>;
  event: Record<string, SignatureData[]>;
};

function constructSearchParams(query: string): [string, string] {
  const hexRe = /^[0-9A-Fa-f]*$/;

  let params = new URLSearchParams();
  params.append('filter', 'false');

  if (query.length === 10 && hexRe.test(query.substring(2))) {
    params.append('function', query);
    return ['lookup', params.toString()];
  } else if (query.length === 66 && hexRe.test(query.substring(2))) {
    params.append('event', query);
    return ['lookup', params.toString()];
  }

  params.append('query', query);
  return ['search', params.toString()];
}

function processSearch(results: SignaturesResponse): React.ReactElement {
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
          </TableRow>
        );
      } else {
        for (let sigResult of sigResults) {
          let isFiltered = sigResult['filtered'] || sigResult['filtered'];

          searchTableRows.push(
            <TableRow
              key={sig + sigType + sigResult['name']}
              sx={{
                '&:last-child td, &:last-child th': { border: 0 },
                backgroundColor: (theme) =>
                  isFiltered ? '#d3d3d3' : '#90ee90',
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
            </TableRow>
          );
        }
      }
    }
  }

  return (
    <Grid item xs={12} md={10} lg={8}>
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
    </Grid>
  );
}

export default function Index() {
  const [query, setQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [searchResults, setSearchResults] = useState(<div></div>);
  const [alertData, setAlertData] = React.useState({
    dismissed: true,
    severity: 'success' as AlertColor,
    message: '',
  });

  const router = useRouter();

  const doSearch = () => {
    const _query = query.trim();
    if (_query.length === 0) {
      setAlertData({
        dismissed: false,
        severity: 'warning',
        message: `Selector is a required field`,
      });
      return;
    }

    setIsSearching(true);
    setAlertData((prevState) => ({
      ...prevState,
      dismissed: true,
    }));
    const [method, params] = constructSearchParams(_query);
    fetch(`${apiEndpoint()}/v1/${method}?${params}`)
      .then((res) => res.json())
      .then((json) => {
        if (json['ok'] === false) {
          throw new Error(json['error']);
        }

        setIsSearching(false);
        setSearchResults(processSearch(json['result'] as SignaturesResponse));
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
    <div>
      <Head>
        <title>Signature Database</title>
        <meta
          name="description"
          content="Search for or import function signatures"
        />
        <meta property="og:type" content="website" />
        <meta property="og:title" content="Signature Database" />
        <meta
          property="og:description"
          content="Search for or import function signatures"
        />
        <meta property="og:image" content="https://openchain.xyz/favicon.png" />
        <meta property="twitter:card" content="summary" />
        <meta property="twitter:title" content="Signature Database" />
        <meta
          property="twitter:description"
          content="Search for or import Ethereum signatures"
        />
        <meta
          property="twitter:url"
          content="https://openchain.xyz/signatures"
        />
        <meta
          property="twitter:image"
          content="https://openchain.xyz/favicon.png"
        />
        <meta property="twitter:site" content="@samczsun" />
        <link rel="icon" href="/favicon.png" />
      </Head>
      <ResponseAppBar router={router} />
      <Grid
        container
        rowSpacing={2}
        alignItems="center"
        justifyContent="center"
        sx={{ mt: '10vh' }}
      >
        <Grid item xs={12}>
          <Box display="flex" justifyContent="center">
            <Typography
              variant="h1"
              align="center"
              fontSize={{
                xs: '2rem',
                sm: '3rem',
                md: '4rem',
                lg: '6rem',
              }}
            >
              Signature Database
            </Typography>
          </Box>
        </Grid>

        <Grid item xs={12}>
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
        </Grid>

        <Grid item xs={8} md={7} lg={6}>
          <Box display="flex" justifyContent="center">
            <TextField
              id="search"
              type="text"
              placeholder="Enter selectors or text to search..."
              fullWidth
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyUp={(event) => {
                if (event.key === 'Enter') {
                  doSearch();
                }
              }}
              autoFocus={true}
            />
          </Box>
        </Grid>

        <Grid item xs={12}>
          <Box display="flex" justifyContent="center">
            <Button
              variant="contained"
              onClick={doSearch}
              disabled={isSearching}
            >
              Search
            </Button>
          </Box>
        </Grid>

        {searchResults}
      </Grid>
    </div>
  );
}
