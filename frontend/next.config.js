/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: true,
    output: 'standalone',
    transpilePackages: ['react-syntax-highlighter', 'swagger-client', 'swagger-ui-react'],
}

module.exports = nextConfig
