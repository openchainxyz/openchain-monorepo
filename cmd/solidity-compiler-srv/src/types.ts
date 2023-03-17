export type CompileRequest = {
    version: string;
    input: StandardInput;
}

export  type StandardInput = {
    language: string;
    sources: Record<string, Source>;
    settings: any;
}

export type Source = {
    content?: string;
}