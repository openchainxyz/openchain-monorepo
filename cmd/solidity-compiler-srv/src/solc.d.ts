declare module "solc/translate" {

    function versionToSemver(version: string): string;

    function translateJsonCompilerOutput(input: any): any;
}

declare module "solc/wrapper" {
    export interface Solc {
        lowlevel: {
            compileStandard: null | function(string): string;
            compileCallback: null | function(string, boolean): string;
            compileMulti: null | function(string, boolean): string;
            compileSingle: null | function(string, boolean): string;
        }
    }

    export default function (input: any): Solc;
}