import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Header from '../components/Header';

interface User {
  id: number;
  username: string;
  email: string;
  phone?: string;
}

interface PasswordForm {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

interface EmailForm {
  new_email: string;
  password: string;
}

interface NotificationSettings {
  email_notifications: boolean;
  push_notifications: boolean;
  order_notifications: boolean;
  marketing_notifications: boolean;
}

export default function Settings() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'account' | 'password' | 'notifications' | 'privacy'>('account');
  const [passwordForm, setPasswordForm] = useState<PasswordForm>({
    current_password: '',
    new_password: '',
    confirm_password: ''
  });
  const [emailForm, setEmailForm] = useState<EmailForm>({
    new_email: '',
    password: ''
  });
  const [notifications, setNotifications] = useState<NotificationSettings>({
    email_notifications: true,
    push_notifications: true,
    order_notifications: true,
    marketing_notifications: false
  });

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
        setEmailForm(prev => ({ ...prev, new_email: data.data.email }));
      } else {
        router.push('/auth/login');
      }
    })
    .catch(() => {
      router.push('/auth/login');
    })
    .finally(() => setLoading(false));
  }, [router]);

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    if (passwordForm.new_password !== passwordForm.confirm_password) {
      alert('Пароли не совпадают');
      return;
    }

    const token = localStorage.getItem('access_token');
    try {
      const response = await fetch('/api/v1/users/password', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          current_password: passwordForm.current_password,
          new_password: passwordForm.new_password
        })
      });

      if (response.ok) {
        alert('Пароль успешно изменен');
        setPasswordForm({
          current_password: '',
          new_password: '',
          confirm_password: ''
        });
      } else {
        const error = await response.json();
        alert(`Ошибка: ${error.message || 'Не удалось изменить пароль'}`);
      }
    } catch (error) {
      alert('Ошибка при изменении пароля');
    }
  };

  const handleEmailChange = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (emailForm.new_email === user?.email) {
      alert('Новый email совпадает с текущим');
      return;
    }

    const token = localStorage.getItem('access_token');
    try {
      const response = await fetch('/api/v1/users/email', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          new_email: emailForm.new_email,
          password: emailForm.password
        })
      });

      if (response.ok) {
        const data = await response.json();
        setUser(prev => prev ? { ...prev, email: data.data.email } : null);
        setEmailForm({ new_email: data.data.email, password: '' });
        alert('Email успешно изменен');
      } else {
        const error = await response.json();
        alert(`Ошибка: ${error.message || 'Не удалось изменить email'}`);
      }
    } catch (error) {
      alert('Ошибка при изменении email');
    }
  };

  const handleNotificationUpdate = async (key: keyof NotificationSettings, value: boolean) => {
    setNotifications(prev => ({
      ...prev,
      [key]: value
    }));
    
    // Здесь будет API запрос для обновления настроек уведомлений
    const token = localStorage.getItem('access_token');
    try {
      await fetch('/api/v1/users/notifications', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          [key]: value
        })
      });
    } catch (error) {
      console.error('Error updating notification settings:', error);
    }
  };



  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header currentPath="/settings" />
        <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-1/4 mb-6"></div>
            <div className="bg-white rounded-lg shadow-sm p-6">
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-4"></div>
              <div className="space-y-4">
                <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                <div className="h-4 bg-gray-200 rounded w-2/3"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header currentPath="/settings" />
      <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Настройки</h1>
          <p className="text-gray-600">Управляйте своим аккаунтом и настройками</p>
        </div>

        <div className="flex space-x-6">
          {/* Боковое меню */}
          <div className="w-64 flex-shrink-0">
            <nav className="bg-white rounded-lg shadow-sm">
              <div className="p-4">
                <ul className="space-y-2">
                  {[
                    { key: 'account', label: 'Аккаунт', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
                    { key: 'password', label: 'Пароль', icon: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z' },
                    { key: 'notifications', label: 'Уведомления', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9' },
                    { key: 'privacy', label: 'Приватность', icon: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' }
                  ].map(item => (
                    <li key={item.key}>
                      <button
                        onClick={() => setActiveTab(item.key as any)}
                        className={`w-full flex items-center px-3 py-2 text-sm font-medium rounded-md ${
                          activeTab === item.key
                            ? 'bg-blue-100 text-blue-700'
                            : 'text-gray-600 hover:bg-gray-100'
                        }`}
                      >
                        <svg className="h-5 w-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={item.icon} />
                        </svg>
                        {item.label}
                      </button>
                    </li>
                  ))}
                </ul>
              </div>
            </nav>
          </div>

          {/* Основной контент */}
          <div className="flex-1">
            <div className="bg-white rounded-lg shadow-sm">
              {activeTab === 'account' && (
                <div className="p-6">
                  <h2 className="text-xl font-semibold text-gray-900 mb-6">Настройки аккаунта</h2>
                  
                  {/* Информация только для чтения */}
                  <div className="mb-8">
                    <h3 className="text-lg font-medium text-gray-900 mb-4">Основная информация</h3>
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Никнейм
                        </label>
                        <div className="flex items-center">
                          <p className="text-gray-900 font-medium">{user?.username}</p>
                          <span className="ml-2 text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded-full">
                            Нельзя изменить
                          </span>
                        </div>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Телефон
                        </label>
                        <div className="flex items-center">
                          <p className="text-gray-900">{user?.phone || 'Не указан'}</p>
                          <span className="ml-2 text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded-full">
                            Нельзя изменить
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Изменение email */}
                  <div>
                    <h3 className="text-lg font-medium text-gray-900 mb-4">Изменить email</h3>
                    <form onSubmit={handleEmailChange} className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Новый email
                        </label>
                        <input
                          type="email"
                          required
                          value={emailForm.new_email}
                          onChange={(e) => setEmailForm({...emailForm, new_email: e.target.value})}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="Введите новый email"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Подтвердите паролем
                        </label>
                        <input
                          type="password"
                          required
                          value={emailForm.password}
                          onChange={(e) => setEmailForm({...emailForm, password: e.target.value})}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          placeholder="Введите текущий пароль"
                        />
                      </div>
                      <div className="flex justify-end">
                        <button type="submit" className="btn btn-primary">
                          Изменить email
                        </button>
                      </div>
                    </form>
                  </div>
                </div>
              )}

              {activeTab === 'password' && (
                <div className="p-6">
                  <h2 className="text-xl font-semibold text-gray-900 mb-6">Изменить пароль</h2>
                  <form onSubmit={handlePasswordChange} className="space-y-6">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Текущий пароль
                      </label>
                      <input
                        type="password"
                        value={passwordForm.current_password}
                        onChange={(e) => setPasswordForm({...passwordForm, current_password: e.target.value})}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Новый пароль
                      </label>
                      <input
                        type="password"
                        value={passwordForm.new_password}
                        onChange={(e) => setPasswordForm({...passwordForm, new_password: e.target.value})}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Подтвердить новый пароль
                      </label>
                      <input
                        type="password"
                        value={passwordForm.confirm_password}
                        onChange={(e) => setPasswordForm({...passwordForm, confirm_password: e.target.value})}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                      />
                    </div>
                    <div className="flex justify-end">
                      <button type="submit" className="btn btn-primary">
                        Изменить пароль
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'notifications' && (
                <div className="p-6">
                  <h2 className="text-xl font-semibold text-gray-900 mb-6">Настройки уведомлений</h2>
                  <div className="space-y-6">
                    {[
                      { key: 'email_notifications', label: 'Email уведомления', description: 'Получать уведомления на email' },
                      { key: 'push_notifications', label: 'Push уведомления', description: 'Получать push уведомления в браузере' },
                      { key: 'order_notifications', label: 'Уведомления о заказах', description: 'Получать уведомления о статусе заказов' },
                      { key: 'marketing_notifications', label: 'Маркетинговые уведомления', description: 'Получать информацию о скидках и акциях' }
                    ].map(item => (
                      <div key={item.key} className="flex items-center justify-between">
                        <div>
                          <h3 className="text-sm font-medium text-gray-900">{item.label}</h3>
                          <p className="text-sm text-gray-500">{item.description}</p>
                        </div>
                        <button
                          onClick={() => handleNotificationUpdate(item.key as keyof NotificationSettings, !notifications[item.key as keyof NotificationSettings])}
                          className={`relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 ${
                            notifications[item.key as keyof NotificationSettings] ? 'bg-blue-600' : 'bg-gray-200'
                          }`}
                        >
                          <span
                            className={`pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200 ${
                              notifications[item.key as keyof NotificationSettings] ? 'translate-x-5' : 'translate-x-0'
                            }`}
                          />
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {activeTab === 'privacy' && (
                <div className="p-6">
                  <h2 className="text-xl font-semibold text-gray-900 mb-6">Приватность и безопасность</h2>
                  <div className="space-y-6">
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                      <h3 className="text-lg font-medium text-blue-900 mb-2">Постоянность аккаунта</h3>
                      <p className="text-sm text-blue-700 mb-4">
                        Ваш аккаунт является постоянным и не может быть удален. Это обеспечивает надежность и стабильность торговой платформы.
                      </p>
                    </div>
                    
                    <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                      <h3 className="text-lg font-medium text-gray-900 mb-2">Экспорт данных</h3>
                      <p className="text-sm text-gray-600 mb-4">
                        Вы можете экспортировать свои данные в формате JSON.
                      </p>
                      <button className="bg-gray-600 text-white px-4 py-2 rounded-md hover:bg-gray-700 transition-colors">
                        Экспортировать данные
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 