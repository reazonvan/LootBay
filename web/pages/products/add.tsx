import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Header from '../../components/Header';

interface Game {
  id: number;
  name: string;
  category: string;
}

interface ProductForm {
  name: string;
  description: string;
  price: number;
  stock: number;
  game_id: number;
  image_url: string;
}

export default function AddProduct() {
  const router = useRouter();
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState<ProductForm>({
    name: '',
    description: '',
    price: 0,
    stock: 1,
    game_id: 0,
    image_url: ''
  });

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/auth/login');
      return;
    }

    // Загружаем список игр
    fetch('/api/v1/games')
      .then(res => res.json())
      .then(data => {
        if (data.success) {
          setGames(data.data || []);
        }
      })
      .catch(error => {
        console.error('Error loading games:', error);
      })
      .finally(() => setLoading(false));
  }, [router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/auth/login');
      return;
    }

    try {
      const response = await fetch('/api/v1/products', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });

      if (response.ok) {
        router.push('/products/my');
      } else {
        const error = await response.json();
        alert(`Ошибка: ${error.message || 'Не удалось создать товар'}`);
      }
    } catch (error) {
      alert('Произошла ошибка при создании товара');
    } finally {
      setSubmitting(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name === 'price' || name === 'stock' || name === 'game_id' ? Number(value) : value
    }));
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header currentPath="/products/add" />
        <div className="max-w-2xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-1/4 mb-6"></div>
            <div className="bg-white rounded-lg shadow-sm p-6">
              <div className="space-y-4">
                <div className="h-4 bg-gray-200 rounded w-1/4"></div>
                <div className="h-10 bg-gray-200 rounded"></div>
                <div className="h-4 bg-gray-200 rounded w-1/4"></div>
                <div className="h-32 bg-gray-200 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header currentPath="/products/add" />
      <div className="max-w-2xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Добавить товар</h1>
          <p className="text-gray-600">Создайте новое объявление для продажи</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow-sm p-6">
          <div className="space-y-6">
            {/* Название товара */}
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                Название товара *
              </label>
              <input
                type="text"
                id="name"
                name="name"
                required
                value={formData.name}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Введите название товара"
              />
            </div>

            {/* Описание */}
            <div>
              <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                Описание *
              </label>
              <textarea
                id="description"
                name="description"
                required
                rows={4}
                value={formData.description}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Опишите ваш товар подробно"
              />
            </div>

            {/* Игра */}
            <div>
              <label htmlFor="game_id" className="block text-sm font-medium text-gray-700 mb-2">
                Игра *
              </label>
              <select
                id="game_id"
                name="game_id"
                required
                value={formData.game_id}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Выберите игру</option>
                {games.map(game => (
                  <option key={game.id} value={game.id}>
                    {game.name} ({game.category})
                  </option>
                ))}
              </select>
            </div>

            {/* Цена и количество */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="price" className="block text-sm font-medium text-gray-700 mb-2">
                  Цена (₽) *
                </label>
                <input
                  type="number"
                  id="price"
                  name="price"
                  required
                  min="1"
                  step="0.01"
                  value={formData.price}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="0"
                />
              </div>
              <div>
                <label htmlFor="stock" className="block text-sm font-medium text-gray-700 mb-2">
                  Количество *
                </label>
                <input
                  type="number"
                  id="stock"
                  name="stock"
                  required
                  min="1"
                  value={formData.stock}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="1"
                />
              </div>
            </div>

            {/* Изображение */}
            <div>
              <label htmlFor="image_url" className="block text-sm font-medium text-gray-700 mb-2">
                Ссылка на изображение
              </label>
              <input
                type="url"
                id="image_url"
                name="image_url"
                value={formData.image_url}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="https://example.com/image.jpg"
              />
              <p className="mt-1 text-sm text-gray-500">
                Добавьте ссылку на изображение товара (необязательно)
              </p>
            </div>

            {/* Предварительный просмотр изображения */}
            {formData.image_url && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Предварительный просмотр
                </label>
                <div className="border border-gray-300 rounded-md p-4">
                  <img
                    src={formData.image_url}
                    alt="Предварительный просмотр"
                    className="w-full h-48 object-cover rounded-md"
                    onError={() => {
                      setFormData(prev => ({ ...prev, image_url: '' }));
                      alert('Не удалось загрузить изображение. Проверьте ссылку.');
                    }}
                  />
                </div>
              </div>
            )}

            {/* Кнопки */}
            <div className="flex justify-end space-x-3 pt-6">
              <button
                type="button"
                onClick={() => router.back()}
                className="btn btn-outline"
                disabled={submitting}
              >
                Отмена
              </button>
              <button
                type="submit"
                className="btn btn-primary"
                disabled={submitting}
              >
                {submitting ? 'Создание...' : 'Создать товар'}
              </button>
            </div>
          </div>
        </form>

        {/* Подсказки */}
        <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-blue-900 mb-2">💡 Советы для успешной продажи</h3>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• Используйте четкое и информативное название</li>
            <li>• Подробно опишите характеристики товара</li>
            <li>• Указывайте честную цену, сравнимую с рыночной</li>
            <li>• Добавляйте качественные изображения</li>
            <li>• Отвечайте быстро на вопросы покупателей</li>
          </ul>
        </div>
      </div>
    </div>
  );
} 