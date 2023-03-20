declare module 'solc/translate' {
    function versionToSemver(version: string): string;

    function translateJsonCompilerOutput(input: any): any;
}

declare module 'solc/wrapper' {
    export type LowLevelAPI = {
        compileStandard: null | ((string) => string);
        compileCallback: null | ((string, boolean) => string);
        compileMulti: null | ((string, boolean) => string);
        compileSingle: null | ((string, boolean) => string);
    };

    export interface Solc {
        lowlevel: LowLevelAPI;
    }

    export default function (input: any): Solc;
}
