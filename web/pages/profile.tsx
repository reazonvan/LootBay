import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/router';
import Header from '../components/Header';

interface User {
  id: number;
  username: string;
  email: string;
  phone?: string;
  balance: number;
  avatar?: string;
  created_at: string;
}

export default function Profile() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/auth/login');
      return;
    }

    fetch('/api/v1/users/profile', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(res => res.json())
    .then(data => {
      if (data.success) {
        setUser(data.data);
      } else {
        router.push('/auth/login');
      }
    })
    .catch(() => {
      router.push('/auth/login');
    })
    .finally(() => setLoading(false));
  }, [router]);

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Проверка типа файла
    if (!file.type.startsWith('image/')) {
      alert('Пожалуйста, выберите изображение');
      return;
    }

    // Проверка размера файла (макс 5MB)
    if (file.size > 5 * 1024 * 1024) {
      alert('Размер файла не должен превышать 5MB');
      return;
    }

    setUploadingAvatar(true);

    try {
      // Создаем FormData для загрузки файла
      const formData = new FormData();
      formData.append('avatar', file);

      const token = localStorage.getItem('access_token');
      const response = await fetch('/api/v1/users/avatar', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      });

      if (response.ok) {
        const data = await response.json();
        setUser(prev => prev ? { ...prev, avatar: data.data.avatar_url } : null);
        alert('Аватар успешно обновлен');
      } else {
        alert('Ошибка при загрузке аватара');
      }
    } catch (error) {
      alert('Произошла ошибка при загрузке аватара');
    } finally {
      setUploadingAvatar(false);
      // Очищаем input для возможности загрузки того же файла снова
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const formatBalance = (balance: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      minimumFractionDigits: 0
    }).format(balance);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header currentPath="/profile" />
        <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
          <div className="animate-pulse">
            <div className="bg-white rounded-lg shadow-sm p-6">
              <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
              <div className="space-y-4">
                <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                <div className="h-4 bg-gray-200 rounded w-2/3"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header currentPath="/profile" />
      <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="bg-white rounded-lg shadow-sm overflow-hidden">
          <div className="bg-gradient-to-r from-blue-600 to-purple-600 px-6 py-8">
            <div className="flex items-center">
              <div className="flex-shrink-0 relative">
                <div 
                  onClick={handleAvatarClick}
                  className="relative cursor-pointer group"
                >
                  {user.avatar ? (
                    <img 
                      src={user.avatar} 
                      alt={user.username}
                      className="h-20 w-20 rounded-full object-cover border-4 border-white"
                    />
                  ) : (
                    <div className="h-20 w-20 rounded-full bg-white flex items-center justify-center border-4 border-white">
                      <span className="text-2xl font-bold text-blue-600">
                        {user.username.charAt(0).toUpperCase()}
                      </span>
                    </div>
                  )}
                  
                  {/* Overlay для загрузки */}
                  <div className="absolute inset-0 bg-black bg-opacity-50 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    {uploadingAvatar ? (
                      <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-white"></div>
                    ) : (
                      <svg className="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                      </svg>
                    )}
                  </div>
                </div>
                
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  onChange={handleAvatarChange}
                  className="hidden"
                />
              </div>
              
              <div className="ml-6">
                <h1 className="text-3xl font-bold text-white">{user.username}</h1>
                <p className="text-blue-100 text-lg">Баланс: {formatBalance(user.balance)}</p>
                <p className="text-blue-100 text-sm">Участник с {formatDate(user.created_at)}</p>
              </div>
            </div>
          </div>

          <div className="px-6 py-6">
            <div className="mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Информация профиля</h2>
              <p className="text-sm text-gray-500 mt-1">
                Некоторые данные можно изменить в настройках
              </p>
            </div>

            <div className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">
                    Никнейм
                  </label>
                  <div className="flex items-center">
                    <p className="text-gray-900 font-medium">{user.username}</p>
                    <span className="ml-2 text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded-full">
                      Нельзя изменить
                    </span>
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">
                    Email
                  </label>
                  <div className="flex items-center">
                    <p className="text-gray-900">{user.email}</p>
                    <button 
                      onClick={() => router.push('/settings')}
                      className="ml-2 text-xs text-blue-600 hover:text-blue-700"
                    >
                      Изменить в настройках
                    </button>
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">
                    Телефон
                  </label>
                  <div className="flex items-center">
                    <p className="text-gray-900">{user.phone || 'Не указан'}</p>
                    <span className="ml-2 text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded-full">
                      Нельзя изменить
                    </span>
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">
                    Баланс
                  </label>
                  <p className="text-gray-900 font-semibold">{formatBalance(user.balance)}</p>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">
                    Дата регистрации
                  </label>
                  <p className="text-gray-900">{formatDate(user.created_at)}</p>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">
                    Пароль
                  </label>
                  <div className="flex items-center">
                    <p className="text-gray-900">••••••••</p>
                    <button 
                      onClick={() => router.push('/settings')}
                      className="ml-2 text-xs text-blue-600 hover:text-blue-700"
                    >
                      Изменить в настройках
                    </button>
                  </div>
                </div>
              </div>
              
              {/* Инструкция по аватару */}
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <h3 className="text-sm font-medium text-blue-900 mb-2">💡 Как изменить аватар</h3>
                <ul className="text-sm text-blue-800 space-y-1">
                  <li>• Нажмите на свой аватар выше</li>
                  <li>• Выберите изображение (JPG, PNG, GIF)</li>
                  <li>• Максимальный размер файла: 5MB</li>
                  <li>• Рекомендуемые размеры: 200x200 пикселей</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 