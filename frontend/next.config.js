/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: true,
    output: 'standalone',
    webpack: (config) => {
        config.resolve.fallback = { ...config.resolve.fallback, net: false, os: false, tls: false, fs: false };
        return config;
    },

    // remove these later
    typescript: {
        ignoreBuildErrors: true,
    },
    eslint: {
        ignoreDuringBuilds: true,
    },
};

module.exports = nextConfig;
