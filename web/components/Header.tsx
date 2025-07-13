import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { AnimatePresence, motion } from 'framer-motion';

interface User {
  id: number;
  username: string;
  email: string;
  balance: number;
  avatar?: string;
  roles?: string[];
}

interface HeaderProps {
  currentPath?: string;
}

export default function Header({ currentPath }: HeaderProps) {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [debugInfo, setDebugInfo] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [showOwnerMenu, setShowOwnerMenu] = useState(false);
  const [showAdminMenu, setShowAdminMenu] = useState(false);

  useEffect(() => {
    setDebugInfo('useEffect запущен, ');
    const token = localStorage.getItem('access_token');
    setDebugInfo(prev => prev + `токен: ${token ? 'есть' : 'нет'}, `);
    if (token) {
      // Извлекаем роли из токена
      let userRoles: string[] = [];
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        userRoles = payload.roles || [];
        setDebugInfo(prev => prev + `роли: ${JSON.stringify(userRoles)}, `);
        
        // Если роли отсутствуют в старом токене, принудительно обновляем
        if (!userRoles || userRoles.length === 0) {
          setDebugInfo(prev => prev + `обновляем токен, `);
          const refreshToken = localStorage.getItem('refresh_token');
          if (refreshToken) {
            // Обновляем токен
            fetch('/api/v1/auth/refresh', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({ refresh_token: refreshToken })
            })
            .then(res => res.json())
            .then(data => {
              if (data.success) {
                localStorage.setItem('access_token', data.data.access_token);
                localStorage.setItem('refresh_token', data.data.refresh_token);
                window.location.reload(); // Перезагружаем страницу
              } else {
                setDebugInfo(prev => prev + `ошибка обновления токена, `);
                // Выходим и предлагаем войти заново
                localStorage.removeItem('access_token');
                localStorage.removeItem('refresh_token');
                window.location.href = '/auth/login';
              }
            })
            .catch(() => {
              setDebugInfo(prev => prev + `catch ошибка, `);
              localStorage.removeItem('access_token');
              localStorage.removeItem('refresh_token');
              window.location.href = '/auth/login';
            });
          } else {
            // Нет refresh токена, выходим
            localStorage.removeItem('access_token');
            window.location.href = '/auth/login';
          }
          return; // Прерываем выполнение
        }
      } catch (error) {
        setDebugInfo(prev => prev + `ошибка парсинга токена, `);
      }

      // Получаем данные пользователя
      fetch('/api/v1/users/profile', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      .then(res => res.json())
      .then(data => {
        if (data.success) {
          setUser({ ...data.data, roles: userRoles });
          setDebugInfo(prev => prev + `user loaded, `);
        }
      })
      .catch(() => {
        // Если токен недействителен, удаляем его
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
      })
      .finally(() => setLoading(false));
    } else {
        setDebugInfo('токен отсутствует');
      setLoading(false);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    setUser(null);
    setShowUserMenu(false);
    router.push('/');
  };

  const formatBalance = (balance: number) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: 'RUB',
      minimumFractionDigits: 0
    }).format(balance);
  };

  // Функции для проверки ролей
  const isOwner = () => user?.roles?.includes('OWNER');
  const isAnyAdmin = () => user?.roles?.some(role => 
    ['ADMIN_GAMES', 'ADMIN_MODERATION', 'ADMIN_SUPPORT'].includes(role)
  );
  const isGamesAdmin = () => user?.roles?.includes('ADMIN_GAMES');
  const isModerationAdmin = () => user?.roles?.includes('ADMIN_MODERATION');
  const isSupportAdmin = () => user?.roles?.includes('ADMIN_SUPPORT');

  // Закрытие всех меню при клике вне их
  const closeAllMenus = () => {
    setShowUserMenu(false);
    setShowOwnerMenu(false);
    setShowAdminMenu(false);
  };

  return (
    <header className="bg-white shadow-soft">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div className="flex items-center">
            <Link href="/" className="text-2xl font-bold text-primary-600">
              LootBay
            </Link>
            {/* Отладочная информация */}
            <div className="ml-4 text-xs bg-red-100 p-1 rounded">
              TEST BUTTON - РОЛИ: {user?.roles?.join(', ') || 'НЕТ РОЛЕЙ'}
            </div>
            {debugInfo && (
              <div className="ml-4 text-xs bg-yellow-100 p-1 rounded max-w-md truncate">
                Debug: {debugInfo}
              </div>
            )}
          </div>
          
          <nav className="hidden md:flex space-x-8">
            <Link 
              href="/products" 
              className={`transition-colors ${
                currentPath === '/products' 
                  ? 'text-primary-600 font-medium' 
                  : 'text-gray-600 hover:text-primary-600'
              }`}
            >
              Товары
            </Link>
            <Link 
              href="/games" 
              className={`transition-colors ${
                currentPath === '/games' 
                  ? 'text-primary-600 font-medium' 
                  : 'text-gray-600 hover:text-primary-600'
              }`}
            >
              Игры
            </Link>
            
            {/* Иконка чата только для авторизованных пользователей */}
            {user && (
              <Link 
                href="/chat" 
                className={`transition-colors flex items-center space-x-1 ${
                  currentPath === '/chat' 
                    ? 'text-primary-600 font-medium' 
                    : 'text-gray-600 hover:text-primary-600'
                }`}
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <span>Чат</span>
              </Link>
            )}
          </nav>

          <div className="flex items-center space-x-4">
            {/* Кнопки ролей */}
            {!loading && user && (
              <>
                {/* Кнопка Владелец */}
                {(true || isOwner()) && (
                  <div className="relative">
                    <button
                      onClick={() => {
                        setShowOwnerMenu(!showOwnerMenu);
                        setShowAdminMenu(false);
                        setShowUserMenu(false);
                      }}
                      className="flex items-center space-x-2 px-3 py-2 text-sm font-medium text-purple-700 bg-purple-50 hover:bg-purple-100 rounded-lg border border-purple-200 transition-all duration-200"
                    >
                      <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                      </svg>
                      <span>Владелец</span>
                      <svg 
                        className={`h-3 w-3 transition-transform ${showOwnerMenu ? 'rotate-180' : ''}`}
                        fill="none" 
                        stroke="currentColor" 
                        viewBox="0 0 24 24"
                      >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                      </svg>
                    </button>

                    <AnimatePresence>
                      {showOwnerMenu && (
                        <motion.div
                          initial={{ opacity: 0, y: -10 }}
                          animate={{ opacity: 1, y: 0 }}
                          exit={{ opacity: 0, y: -10 }}
                          transition={{ duration: 0.2 }}
                          className="absolute left-0 mt-2 w-56 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50"
                        >
                          <Link
                            href="/admin"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 transition-colors"
                            onClick={closeAllMenus}
                          >
                            <div className="flex items-center">
                              <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                              </svg>
                              Админ-панель
                            </div>
                          </Link>
                          <Link
                            href="/admin/roles"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 transition-colors"
                            onClick={closeAllMenus}
                          >
                            <div className="flex items-center">
                              <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a1.5 1.5 0 01-3 0V5.197m3 4.803a24.056 24.056 0 01-5.77-3.597m0 0a23.906 23.906 0 00-5.115 3.597m9.885 0A23.906 23.906 0 0015.115 9.4M20.5 10V5.197m0 4.803v4.096m-16-4.096V5.197m0 4.803a24.056 24.056 0 005.77-3.597m0 0a23.906 23.906 0 005.115 3.597m-9.885 0A23.906 23.906 0 008.885 9.4M3.5 10V5.197M3.5 14.103v4.096" />
                              </svg>
                              Управление ролями
                            </div>
                          </Link>
                          <Link
                            href="/admin/statistics"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 transition-colors"
                            onClick={closeAllMenus}
                          >
                            <div className="flex items-center">
                              <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                              </svg>
                              Статистика
                            </div>
                          </Link>
                          <hr className="my-1" />
                          <Link
                            href="/admin/games"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 transition-colors"
                            onClick={closeAllMenus}
                          >
                            <div className="flex items-center">
                              <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                              </svg>
                              Управление играми
                            </div>
                          </Link>
                          <Link
                            href="/admin/moderation"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 transition-colors"
                            onClick={closeAllMenus}
                          >
                            <div className="flex items-center">
                              <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                              </svg>
                              Модерация
                            </div>
                          </Link>
                          <Link
                            href="/admin/support"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 transition-colors"
                            onClick={closeAllMenus}
                          >
                            <div className="flex items-center">
                              <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192L5.636 18.364M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
                              </svg>
                              Поддержка
                            </div>
                          </Link>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                )}

                {/* Кнопка Админ (только если не владелец) */}
                {!isOwner() && isAnyAdmin() && (
                  <div className="relative">
                    <button
                      onClick={() => {
                        setShowAdminMenu(!showAdminMenu);
                        setShowOwnerMenu(false);
                        setShowUserMenu(false);
                      }}
                      className="flex items-center space-x-2 px-3 py-2 text-sm font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg border border-blue-200 transition-all duration-200"
                    >
                      <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                      </svg>
                      <span>Админ</span>
                      <svg 
                        className={`h-3 w-3 transition-transform ${showAdminMenu ? 'rotate-180' : ''}`}
                        fill="none" 
                        stroke="currentColor" 
                        viewBox="0 0 24 24"
                      >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                      </svg>
                    </button>

                    <AnimatePresence>
                      {showAdminMenu && (
                        <motion.div
                          initial={{ opacity: 0, y: -10 }}
                          animate={{ opacity: 1, y: 0 }}
                          exit={{ opacity: 0, y: -10 }}
                          transition={{ duration: 0.2 }}
                          className="absolute left-0 mt-2 w-56 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50"
                        >
                          {isGamesAdmin() && (
                            <>
                              <Link
                                href="/admin/games"
                                className="block px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 transition-colors"
                                onClick={closeAllMenus}
                              >
                                <div className="flex items-center">
                                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                                  </svg>
                                  Управление играми
                                </div>
                              </Link>
                              <Link
                                href="/admin/categories"
                                className="block px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 transition-colors"
                                onClick={closeAllMenus}
                              >
                                <div className="flex items-center">
                                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                  </svg>
                                  Категории
                                </div>
                              </Link>
                            </>
                          )}
                          
                          {isModerationAdmin() && (
                            <>
                              <Link
                                href="/admin/moderation"
                                className="block px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 transition-colors"
                                onClick={closeAllMenus}
                              >
                                <div className="flex items-center">
                                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                  </svg>
                                  Модерация контента
                                </div>
                              </Link>
                              <Link
                                href="/admin/reports"
                                className="block px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 transition-colors"
                                onClick={closeAllMenus}
                              >
                                <div className="flex items-center">
                                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                  </svg>
                                  Жалобы
                                </div>
                              </Link>
                            </>
                          )}
                          
                          {isSupportAdmin() && (
                            <>
                              <Link
                                href="/admin/support"
                                className="block px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 transition-colors"
                                onClick={closeAllMenus}
                              >
                                <div className="flex items-center">
                                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192L5.636 18.364M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
                                  </svg>
                                  Поддержка пользователей
                                </div>
                              </Link>
                              <Link
                                href="/admin/tickets"
                                className="block px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 transition-colors"
                                onClick={closeAllMenus}
                              >
                                <div className="flex items-center">
                                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" />
                                  </svg>
                                  Тикеты поддержки
                                </div>
                              </Link>
                            </>
                          )}
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                )}
              </>
            )}

            {loading ? (
              <div className="animate-pulse">
                <div className="h-8 w-32 bg-gray-200 rounded"></div>
              </div>
            ) : user ? (
              <div className="relative">
                <button
                  onClick={() => setShowUserMenu(!showUserMenu)}
                  className="flex items-center space-x-3 text-gray-700 hover:text-primary-600 transition-colors"
                >
                  <div className="flex items-center space-x-2">
                    {user.avatar ? (
                      <img 
                        src={user.avatar} 
                        alt={user.username}
                        className="h-8 w-8 rounded-full object-cover"
                      />
                    ) : (
                      <div className="h-8 w-8 rounded-full bg-primary-600 flex items-center justify-center">
                        <span className="text-white text-sm font-medium">
                          {user.username.charAt(0).toUpperCase()}
                        </span>
                      </div>
                    )}
                    <div className="text-left">
                      <p className="text-sm font-medium">{user.username}</p>
                      <p className="text-xs text-gray-500">{formatBalance(user.balance)}</p>
                    </div>
                  </div>
                  <svg 
                    className={`h-4 w-4 transition-transform ${showUserMenu ? 'rotate-180' : ''}`}
                    fill="none" 
                    stroke="currentColor" 
                    viewBox="0 0 24 24"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                  </svg>
                </button>

                <AnimatePresence>
                  {showUserMenu && (
                    <motion.div
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      transition={{ duration: 0.2 }}
                      className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50"
                    >
                      <Link
                        href="/profile"
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        onClick={() => setShowUserMenu(false)}
                      >
                        <div className="flex items-center">
                          <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                          </svg>
                          Профиль
                        </div>
                      </Link>
                      
                      <Link
                        href="/orders"
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        onClick={() => setShowUserMenu(false)}
                      >
                        <div className="flex items-center">
                          <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                          </svg>
                          Мои заказы
                        </div>
                      </Link>
                      
                      <Link
                        href="/products/my"
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        onClick={() => setShowUserMenu(false)}
                      >
                        <div className="flex items-center">
                          <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                          </svg>
                          Мои товары
                        </div>
                      </Link>
                      
                      <Link
                        href="/wallet"
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        onClick={() => setShowUserMenu(false)}
                      >
                        <div className="flex items-center">
                          <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                          </svg>
                          Кошелек
                        </div>
                      </Link>
                      
                      <hr className="my-1" />
                      
                      <Link
                        href="/settings"
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        onClick={() => setShowUserMenu(false)}
                      >
                        <div className="flex items-center">
                          <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                          </svg>
                          Настройки
                        </div>
                      </Link>


                      
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2 text-sm text-red-700 hover:bg-red-50 transition-colors"
                      >
                        <div className="flex items-center">
                          <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                          </svg>
                          Выйти
                        </div>
                      </button>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ) : (
              <>
                <Link href="/auth/login" className="btn btn-outline btn-sm">
                  Войти
                </Link>
                <Link href="/auth/register" className="btn btn-primary btn-sm">
                  Регистрация
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
} 