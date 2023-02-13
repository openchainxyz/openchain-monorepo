import Navbar from '@components/Navbar';
import { DecodeTab } from '@components/tools/abi/DecodeTab';
import { EncodeTab } from '@components/tools/abi/EncodeTab';
import { TabContext, TabPanel } from '@mui/lab';
import { Tab, Tabs } from '@mui/material';

import Box from '@mui/material/Box';
import React, { useState } from 'react';

export default function Index() {
    const [tab, setTab] = useState<string>('decode');

    return (
        <Navbar
            title={'ABI Tools'}
            description={'Handy tools for ABI-encoded data'}
            icon={'/abi-tools.png'}
            url={'/tools/abi'}
            content={
                <>
                    <TabContext value={tab}>
                        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                            <Tabs value={tab} onChange={(event, newValue) => setTab(newValue)}>
                                <Tab value="decode" label="Decode" />
                                <Tab value="encode" label="Encode" />
                            </Tabs>
                        </Box>
                        <TabPanel value={'decode'}>
                            <DecodeTab />
                        </TabPanel>
                        <TabPanel value={'encode'}>
                            <EncodeTab />
                        </TabPanel>
                    </TabContext>
                </>
            }
        />
    );
}
