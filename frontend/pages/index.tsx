import Grid2 from '@mui/material/Unstable_Grid2';
import Image from 'next/image';
import Link from 'next/link';
import * as React from 'react';
import Navbar from '../components/Navbar';

type Service = {
    icon: string;
    name: string;
    path: string;
    description: string;
};

const services: Service[] = [
    {
        icon: '/signatures.png',
        name: 'Signature Database',
        path: '/signatures',
        description:
            'Look up unknown function selectors or event topics, or contribute your own data for others to use',
    },
    {
        icon: '/tracer.png',
        name: 'Transaction Tracer',
        path: '/trace',
        description:
            'Want to dig deep into a EVM-compatible transaction? This easy-to-use transaction tracer is for you.',
    },
    {
        icon: '/abi-tools.png',
        name: 'ABI Tools',
        path: '/tools/abi',
        description: 'Some handy tools for encoding/decoding ABI data',
    },
];
export default function Home() {
    const serviceBoxes = services.map((service, idx) => {
        return (
            <Grid2 display={'flex'} justifyContent="center" alignItems="center" key={idx} xs={6}>
                <Grid2 container direction={'column'}>
                    <Link href={service.path}>
                        <Grid2 display="flex" p={2} gap={2} justifyContent="center" alignItems="center">
                            <Image src={service.icon} width="24" height="24" alt="logo" />
                            {service.name}
                        </Grid2>
                    </Link>
                    <Grid2>{service.description}</Grid2>
                </Grid2>
            </Grid2>
        );
    });

    return (
        <Navbar
            title={'OpenChain'}
            description={'OpenChain'}
            icon={'/openchain.png'}
            url={'/'}
            content={
                <Grid2 container p={2}>
                    {serviceBoxes}
                </Grid2>
            }
        />
    );
}
