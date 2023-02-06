import { DarkMode, Description, GitHub, LightMode, Twitter } from '@mui/icons-material';
import { Container, Divider, Typography } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2';
import { Box } from '@mui/system';
import Head from 'next/head';
import Image from 'next/image';
import Link from 'next/link';
import * as React from 'react';
import { useContext } from 'react';
import { DarkModeContext } from './DarkModeContext';

export type NavbarLink = {
    name: string;
    url: string;
};

export type NavbarProps = {
    title: string;
    description: string;
    icon: string;
    url: string;
    content: React.ReactNode;
    links?: NavbarLink[];
};

function Navbar(props: NavbarProps) {
    const [usingDarkMode, setUsingDarkMode] = useContext(DarkModeContext);

    return (
        <div>
            <Head>
                <title>{props.title}</title>
                <meta name="description" content={props.description} />
                <meta property="og:type" content="website" />
                <meta property="og:title" content={props.title} />
                <meta property="og:description" content={props.description} />
                <meta property="og:image" content={props.icon} />
                <meta property="twitter:card" content="summary" />
                <meta property="twitter:title" content={props.title} />
                <meta property="twitter:description" content={props.description} />
                <meta property="twitter:url" content={'https://openchain.xyz' + props.url} />
                <meta property="twitter:image" content={props.icon} />
                <meta property="twitter:site" content="@samczsun" />
                <link rel="icon" href={props.icon} />
            </Head>

            <Container maxWidth={'md'}>
                <Grid2 container justifyContent="center" alignContent="center" p={2} spacing={1}>
                    <Grid2 style={{ cursor: 'pointer' }}>
                        <Link href={props.url}>
                            <Box>
                                <Image src={props.icon} width="24" height="24" alt="logo" />
                            </Box>
                        </Link>
                    </Grid2>
                    <Grid2 sx={{ display: { xs: 'none', md: 'initial' }, marginRight: 2 }}>
                        <Link href={props.url}>
                            <Typography fontFamily="NBInter">{props.title}</Typography>
                        </Link>
                    </Grid2>
                    {(props.links || []).map((link, idx) => {
                        const content = <Typography fontFamily="NBInter">{link.name}</Typography>;

                        if (link.url.startsWith('/')) {
                            return (
                                <Grid2 key={idx}>
                                    <Link href={link.url}>{content}</Link>
                                </Grid2>
                            );
                        } else {
                            return (
                                <Grid2 key={idx}>
                                    <a href={link.url}>{content}</a>
                                </Grid2>
                            );
                        }
                    })}
                    <Grid2 xs></Grid2>
                    <Grid2>
                        <a
                            href="https://github.com/openchainxyz/openchain-monorepo"
                            target={'_blank'}
                            rel={'noreferrer noopener'}
                        >
                            <GitHub />
                        </a>
                    </Grid2>
                    <Grid2>
                        <a href="https://docs.openchain.xyz" target={'_blank'} rel={'noreferrer noopener'}>
                            <Description />
                        </a>
                    </Grid2>
                    <Grid2>
                        <a href="https://twitter.com/samczsun" target={'_blank'} rel={'noreferrer noopener'}>
                            <Twitter />
                        </a>
                    </Grid2>
                    <Grid2 onClick={() => setUsingDarkMode(!usingDarkMode)}>
                        {usingDarkMode ? <LightMode /> : <DarkMode />}
                    </Grid2>
                </Grid2>
                <Divider></Divider>
                {props.content}
            </Container>
        </div>
    );
}

export default Navbar;
