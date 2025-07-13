import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Header from '../components/Header';

interface User {
  id: string;
  username: string;
  email: string;
  phone?: string;
  first_name?: string;
  last_name?: string;
  is_active: boolean;
  created_at: string;
  roles: Role[];
}

interface Role {
  id: string;
  name: string;
  display_name: string;
  description: string;
  level: number;
}

interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description: string;
}

export default function AdminPanel() {
  const router = useRouter();
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'users' | 'roles' | 'permissions'>('users');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // Проверка авторизации и роли OWNER
  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/auth/login');
      return;
    }

    // Проверяем роль OWNER из токена
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const userRoles = payload.roles || [];
      if (!userRoles.includes('OWNER')) {
        setMessage({ type: 'error', text: 'Доступ запрещен. Только владелец может управлять ролями.' });
        setTimeout(() => router.push('/'), 3000);
        return;
      }
    } catch (error) {
      router.push('/auth/login');
      return;
    }

    fetchData();
  }, [router]);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('access_token');
      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      // Загружаем пользователей, роли и разрешения параллельно
      const [usersRes, rolesRes, permissionsRes] = await Promise.all([
        fetch('/api/v1/roles/users', { headers }),
        fetch('/api/v1/roles', { headers }),
        fetch('/api/v1/roles/permissions', { headers })
      ]);

      if (usersRes.ok) {
        const usersData = await usersRes.json();
        setUsers(usersData.data || []);
      }

      if (rolesRes.ok) {
        const rolesData = await rolesRes.json();
        setRoles(rolesData.data || []);
      }

      if (permissionsRes.ok) {
        const permissionsData = await permissionsRes.json();
        setPermissions(permissionsData.data || []);
      }

    } catch (error) {
      console.error('Ошибка загрузки данных:', error);
      setMessage({ type: 'error', text: 'Ошибка загрузки данных' });
    } finally {
      setLoading(false);
    }
  };

  const assignRole = async (userId: string, roleName: string) => {
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch(`/api/v1/roles/users/${userId}/roles`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ role_name: roleName })
      });

      if (response.ok) {
        setMessage({ type: 'success', text: `Роль ${roleName} назначена пользователю` });
        fetchData(); // Перезагружаем данные
      } else {
        const error = await response.json();
        setMessage({ type: 'error', text: error.message || 'Ошибка назначения роли' });
      }
    } catch (error) {
      console.error('Ошибка назначения роли:', error);
      setMessage({ type: 'error', text: 'Ошибка назначения роли' });
    }
  };

  const removeRole = async (userId: string, roleName: string) => {
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch(`/api/v1/roles/users/${userId}/roles`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ role_name: roleName })
      });

      if (response.ok) {
        setMessage({ type: 'success', text: `Роль ${roleName} снята с пользователя` });
        fetchData(); // Перезагружаем данные
      } else {
        const error = await response.json();
        setMessage({ type: 'error', text: error.message || 'Ошибка снятия роли' });
      }
    } catch (error) {
      console.error('Ошибка снятия роли:', error);
      setMessage({ type: 'error', text: 'Ошибка снятия роли' });
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getRoleLevel = (roleName: string) => {
    const role = roles.find(r => r.name === roleName);
    return role ? role.level : 0;
  };

  const getRoleDisplayName = (roleName: string) => {
    const role = roles.find(r => r.name === roleName);
    return role ? role.display_name : roleName;
  };

  const getRoleColor = (roleName: string) => {
    const colors: { [key: string]: string } = {
      'OWNER': 'bg-red-100 text-red-800',
      'ADMIN_GAMES': 'bg-purple-100 text-purple-800',
      'ADMIN_MODERATION': 'bg-orange-100 text-orange-800',
      'ADMIN_SUPPORT': 'bg-blue-100 text-blue-800',
      'USER': 'bg-gray-100 text-gray-800'
    };
    return colors[roleName] || 'bg-gray-100 text-gray-800';
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header />
        <div className="flex justify-center items-center h-96">
          <div className="text-xl text-gray-600">Загрузка...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Админ-панель</h1>
          <p className="mt-2 text-gray-600">Управление пользователями и ролями системы</p>
        </div>

        {/* Сообщения */}
        {message && (
          <div className={`mb-6 p-4 rounded-lg ${
            message.type === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
          }`}>
            {message.text}
          </div>
        )}

        {/* Табы */}
        <div className="border-b border-gray-200 mb-8">
          <nav className="-mb-px flex space-x-8">
            {[
              { key: 'users', label: 'Пользователи' },
              { key: 'roles', label: 'Роли' },
              { key: 'permissions', label: 'Разрешения' }
            ].map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key as any)}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.key
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Содержимое табов */}
        {activeTab === 'users' && (
          <div className="bg-white shadow overflow-hidden sm:rounded-md">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Пользователи ({users.length})
              </h3>
              
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Пользователь
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Email
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Роли
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Статус
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Дата регистрации
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Действия
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {users.map((user) => (
                      <tr key={user.id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex items-center">
                            <div className="flex-shrink-0 h-10 w-10">
                              <div className="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                                <span className="text-sm font-medium text-gray-700">
                                  {user.username[0].toUpperCase()}
                                </span>
                              </div>
                            </div>
                            <div className="ml-4">
                              <div className="text-sm font-medium text-gray-900">{user.username}</div>
                              {user.first_name && user.last_name && (
                                <div className="text-sm text-gray-500">
                                  {user.first_name} {user.last_name}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {user.email}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex flex-wrap gap-1">
                            {user.roles?.map((role) => (
                              <span
                                key={role.name}
                                className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleColor(role.name)}`}
                              >
                                {getRoleDisplayName(role.name)}
                              </span>
                            ))}
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                          }`}>
                            {user.is_active ? 'Активен' : 'Заблокирован'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {formatDate(user.created_at)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                          <button
                            onClick={() => setSelectedUser(user)}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            Управление ролями
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'roles' && (
          <div className="bg-white shadow overflow-hidden sm:rounded-md">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Роли системы ({roles.length})
              </h3>
              
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {roles.map((role) => (
                  <div key={role.id} className="border border-gray-200 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="text-lg font-medium text-gray-900">{role.display_name}</h4>
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleColor(role.name)}`}>
                        Уровень {role.level}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 mb-3">{role.description}</p>
                    <div className="text-xs text-gray-500">
                      Системное имя: <code className="bg-gray-100 px-1 rounded">{role.name}</code>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'permissions' && (
          <div className="bg-white shadow overflow-hidden sm:rounded-md">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Разрешения ({permissions.length})
              </h3>
              
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Название
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Ресурс
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Действие
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Описание
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {permissions.map((permission) => (
                      <tr key={permission.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          {permission.name}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          <code className="bg-gray-100 px-2 py-1 rounded text-xs">{permission.resource}</code>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          <code className="bg-gray-100 px-2 py-1 rounded text-xs">{permission.action}</code>
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {permission.description}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Модальное окно управления ролями */}
      {selectedUser && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="mt-3">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Управление ролями: {selectedUser.username}
              </h3>
              
              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Текущие роли:</h4>
                <div className="flex flex-wrap gap-2 mb-4">
                  {selectedUser.roles?.map((role) => (
                    <div key={role.name} className="flex items-center gap-1">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleColor(role.name)}`}>
                        {getRoleDisplayName(role.name)}
                      </span>
                      {role.name !== 'USER' && role.name !== 'OWNER' && (
                        <button
                          onClick={() => removeRole(selectedUser.id, role.name)}
                          className="text-red-600 hover:text-red-800 text-xs"
                          title="Снять роль"
                        >
                          ×
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Назначить роль:</h4>
                <div className="space-y-2">
                  {roles
                    .filter(role => role.name !== 'USER' && role.name !== 'OWNER')
                    .filter(role => !selectedUser.roles?.some(userRole => userRole.name === role.name))
                    .map((role) => (
                      <button
                        key={role.id}
                        onClick={() => assignRole(selectedUser.id, role.name)}
                        className="w-full text-left px-3 py-2 border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        <div className="font-medium">{role.display_name}</div>
                        <div className="text-sm text-gray-500">{role.description}</div>
                      </button>
                    ))}
                </div>
              </div>

              <div className="flex justify-end space-x-3">
                <button
                  onClick={() => setSelectedUser(null)}
                  className="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Закрыть
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
} 