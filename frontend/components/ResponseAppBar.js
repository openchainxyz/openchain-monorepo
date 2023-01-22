import * as React from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Menu from '@mui/material/Menu';
import Container from '@mui/material/Container';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';

import PropTypes from 'prop-types';
import useScrollTrigger from '@mui/material/useScrollTrigger';

const pages = ['home', 'import', 'docs'];

function ElevationScroll(props) {
    const {children} = props;
    // Note that you normally won't need to set the window ref as useScrollTrigger
    // will default to window.
    // This is only being set here because the demo is in an iframe.
    const trigger = useScrollTrigger({
        disableHysteresis: true,
        threshold: 0,
    });

    return React.cloneElement(children, {
        elevation: trigger ? 4 : 0,
    });
}

ElevationScroll.propTypes = {
    children: PropTypes.element.isRequired,
};


const ResponsiveAppBar = ({router}) => {
    const generateOnClickHandler = page => {
        return () => {
            if (page === 'docs') {
                window.location = 'https://docs.openchain.xyz'
            } else {
                if (page === 'home') page = '';
                router.push(`/signatures/${page}`);
            }
        };
    }

    return (
        <ElevationScroll>
            <AppBar position="sticky" color="inherit">
                <Container maxWidth="xl">
                    <Toolbar disableGutters>
                        <Box sx={{flexGrow: 1, display: 'flex'}}>
                            {pages.map((page) => (
                                <Button
                                    key={page}
                                    onClick={generateOnClickHandler(page)}
                                    sx={{my: 2, color: 'black', display: 'block'}}
                                >
                                    {page}
                                </Button>
                            ))}
                        </Box>
                    </Toolbar>
                </Container>
            </AppBar>
        </ElevationScroll>
    );
};
export default ResponsiveAppBar;