/** @type {import('next').NextConfig} */
const nextConfig = {
  // Добавляем standalone режим для Docker
  output: 'standalone',
  
  // Оптимизация для разработки
  experimental: {
    // Отключаем тяжелые experimental features
    esmExternals: false,
  },
  
  // Быстрая компиляция в development
  swcMinify: false, // Отключаем минификацию в dev
  
  // Оптимизация сборки
  webpack: (config, { dev, isServer }) => {
    if (dev) {
      // Уменьшаем количество chunks в development
      config.optimization.splitChunks = {
        chunks: 'all',
        cacheGroups: {
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            chunks: 'all',
          },
        },
      };
      
      // Отключаем source maps для ускорения
      config.devtool = false;
      
      // Уменьшаем babel processing
      config.module.rules.forEach((rule) => {
        if (rule.use && rule.use.loader === 'next-babel-loader') {
          rule.use.options.cacheDirectory = true;
        }
      });
    }
    
    return config;
  },
  
  // Отключаем telemetry
  telemetry: false,
  
  // Быстрая hot reload
  reactStrictMode: false, // Отключаем strict mode для ускорения

  // API routes configuration
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://api-gateway:8080/api/:path*',
      },
    ];
  },

  // Image optimization
  images: {
    domains: ['localhost', 'lootbay.com'],
    formats: ['image/webp', 'image/avif'],
  },

  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080',
  },

  // Headers for security
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
        ],
      },
    ];
  },
};

module.exports = nextConfig; 