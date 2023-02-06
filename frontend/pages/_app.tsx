// noinspection ES6UnusedImports
import {} from '@mui/lab/themeAugmentation';
import { createTheme, CssBaseline, ThemeProvider, useMediaQuery } from '@mui/material';

import { useLocalStorageValue } from '@react-hookz/web';
import * as React from 'react';
import { DarkModeContext } from '../components/DarkModeContext';
import '../styles/globals.css';

function MyApp({ Component, pageProps }: { Component: any; pageProps: any }) {
    const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)');
    const { value: localStorageDarkMode, set: setLocalStorageDarkMode } = useLocalStorageValue<boolean>('pref:dark', {
        initializeWithValue: false,
    });

    const useDarkMode = localStorageDarkMode === undefined ? prefersDarkMode : localStorageDarkMode;

    React.useEffect(() => {
        document.documentElement.setAttribute('data-theme', useDarkMode ? 'dark' : 'light');
    }, [useDarkMode]);

    const theme = React.useMemo(() => {
        return createTheme({
            palette: {
                mode: useDarkMode ? 'dark' : 'light',
            },
            components: {
                MuiDialogTitle: {
                    styleOverrides: {
                        root: {
                            paddingBottom: '6px',
                        },
                    },
                },
                MuiDialogContent: {
                    styleOverrides: {
                        root: {
                            paddingTop: '6px',
                        },
                    },
                },
                MuiTreeView: {
                    styleOverrides: {
                        root: {
                            // disabling this for now - if the tree is responsive then the scrollbar is at the bottom of the trace
                            // this makes it really annoying to scroll left/right if the trace is super long, because you have to go
                            // all the way down to the scrollbar
                            // overflow: 'auto',
                            // paddingBottom: '15px', // so the scrollbar doesn't cover the last trace item
                        },
                    },
                },
                MuiTreeItem: {
                    styleOverrides: {
                        content: {
                            cursor: 'initial',
                        },
                        label: {
                            fontSize: 'initial',
                        },
                        iconContainer: {
                            cursor: 'pointer',
                        },
                    },
                },
                MuiDialog: {
                    styleOverrides: {
                        root: {
                            pointerEvents: 'none',
                        },
                    },
                },
                MuiTypography: {
                    styleOverrides: {
                        h5: {
                            fontFamily: 'monospace',
                            fontSize: 'initial',
                            whiteSpace: 'nowrap',
                        },
                        h6: {
                            fontFamily: 'NBInter',
                        },
                        body1: {
                            fontFamily: 'monospace',
                            wordWrap: 'break-word',
                            whiteSpace: 'break-spaces',
                        },
                        body2: {
                            fontFamily: 'monospace',
                            letterSpacing: 'initial',
                        },
                    },
                },
                MuiTableCell: {
                    styleOverrides: {
                        root: {
                            padding: '0px 16px',
                            fontFamily: 'monospace',
                            letterSpacing: 'initial',
                            fontSize: '13px',
                        },
                        // head: {
                        //     fontFamily: 'monospace',
                        //     letterSpacing: 'initial',
                        // },
                        // body: {
                        // },
                    },
                },
            },
        });
    }, [useDarkMode]);

    return (
        <>
            <DarkModeContext.Provider
                value={[
                    useDarkMode,
                    (newDarkMode) => {
                        if (typeof newDarkMode === 'boolean') {
                            setLocalStorageDarkMode(newDarkMode);
                        } else {
                            setLocalStorageDarkMode(newDarkMode(useDarkMode));
                        }
                    },
                ]}
            >
                <ThemeProvider theme={theme}>
                    <CssBaseline />
                    <Component {...pageProps} />
                </ThemeProvider>
            </DarkModeContext.Provider>
        </>
    );
}

export default MyApp;
