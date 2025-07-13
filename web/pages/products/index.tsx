import { useState, useEffect } from 'react';
import Head from 'next/head';
import Link from 'next/link';
import { motion } from 'framer-motion';
import Header from '../../components/Header';

interface Product {
  id: number;
  title: string;
  price: number;
  category: string;
  game: string;
  image?: string;
  seller: string;
  description: string;
  type: 'account' | 'currency' | 'item' | 'service';
  status: 'active' | 'sold' | 'pending';
}

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await fetch('/api/v1/products');
        if (response.ok) {
          const data = await response.json();
          setProducts(data.data || []);
        } else {
          setError('Ошибка загрузки товаров');
        }
      } catch (err) {
        setError('Ошибка соединения');
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    // Здесь будет логика поиска
    console.log('Поиск:', searchQuery);
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB'
    }).format(price);
  };

  const getTypeLabel = (type: string) => {
    const labels = {
      account: 'Аккаунт',
      currency: 'Валюта',
      item: 'Предмет',
      service: 'Услуга'
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getStatusColor = (status: string) => {
    const colors = {
      active: 'bg-green-100 text-green-800',
      sold: 'bg-red-100 text-red-800',
      pending: 'bg-yellow-100 text-yellow-800'
    };
    return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800';
  };

  return (
    <>
      <Head>
        <title>Товары - LootBay</title>
        <meta name="description" content="Каталог игровых товаров на LootBay" />
      </Head>

      <div className="min-h-screen bg-gray-50">
        <Header currentPath="/products" />

        {/* Page Content */}
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Search and Filters */}
          <div className="mb-8">
            <form onSubmit={handleSearch} className="flex gap-4 mb-6">
              <div className="flex-1">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Поиск товаров..."
                  className="input"
                />
              </div>
              <button type="submit" className="btn btn-primary">
                Найти
              </button>
            </form>

            <div className="flex flex-wrap gap-4">
              <select className="input w-auto">
                <option value="">Все игры</option>
                <option value="dota2">Dota 2</option>
                <option value="csgo">CS:GO</option>
                <option value="pubg">PUBG</option>
              </select>
              
              <select className="input w-auto">
                <option value="">Все типы</option>
                <option value="account">Аккаунты</option>
                <option value="currency">Валюта</option>
                <option value="item">Предметы</option>
                <option value="service">Услуги</option>
              </select>
              
              <select className="input w-auto">
                <option value="">Сортировка</option>
                <option value="price_asc">Цена по возрастанию</option>
                <option value="price_desc">Цена по убыванию</option>
                <option value="created_desc">Сначала новые</option>
              </select>
            </div>
          </div>

          {/* Products Grid */}
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[...Array(6)].map((_, i) => (
                <div key={i} className="card animate-pulse">
                  <div className="h-48 bg-gray-200 rounded-t-lg"></div>
                  <div className="card-body space-y-3">
                    <div className="h-4 bg-gray-200 rounded"></div>
                    <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                    <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                  </div>
                </div>
              ))}
            </div>
          ) : error ? (
            <div className="text-center py-12">
              <p className="text-red-600">{error}</p>
            </div>
          ) : products.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-gray-600">Товары не найдены</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {products.map((product) => (
                <motion.div
                  key={product.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5 }}
                  className="card hover:shadow-lg transition-shadow"
                >
                  <div className="relative">
                    <img
                      src={product.image || '/placeholder-product.jpg'}
                      alt={product.title}
                      className="w-full h-48 object-cover rounded-t-lg"
                    />
                    <div className="absolute top-2 right-2">
                      <span className={`badge ${getStatusColor(product.status)}`}>
                        {product.status}
                      </span>
                    </div>
                  </div>
                  
                  <div className="card-body">
                    <div className="flex justify-between items-start mb-2">
                      <h3 className="text-lg font-semibold text-gray-900 line-clamp-1">
                        {product.title}
                      </h3>
                      <span className="text-lg font-bold text-primary-600">
                        {formatPrice(product.price)}
                      </span>
                    </div>
                    
                    <p className="text-gray-600 text-sm mb-3 line-clamp-2">
                      {product.description}
                    </p>
                    
                    <div className="flex justify-between items-center mb-3">
                      <span className="text-sm text-gray-500">
                        {product.game} • {getTypeLabel(product.type)}
                      </span>
                      <span className="text-sm text-gray-500">
                        {product.seller}
                      </span>
                    </div>
                    
                    <Link
                      href={`/products/${product.id}`}
                      className="btn btn-primary w-full btn-sm"
                    >
                      Подробнее
                    </Link>
                  </div>
                </motion.div>
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  );
} 