export function apiEndpoint() {
    return ['localhost', '127.0.0.1'].indexOf(window.location.hostname) !== -1 && false
        ? 'http://localhost:34887'
        : 'https://api.openchain.xyz/signature-database';
}
