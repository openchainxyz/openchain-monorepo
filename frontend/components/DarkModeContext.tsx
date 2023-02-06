import React, { createContext, Dispatch, SetStateAction } from 'react';

type DarkModeContextType = [boolean, Dispatch<SetStateAction<boolean>>];

export const DarkModeContext = createContext([false, () => {}] as DarkModeContextType);
