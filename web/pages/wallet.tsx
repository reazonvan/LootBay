import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Header from '../components/Header';

interface Transaction {
  id: number;
  user_id: number;
  type: 'deposit' | 'withdraw' | 'purchase' | 'sale';
  amount: number;
  description: string;
  status: 'pending' | 'completed' | 'failed';
  created_at: string;
}

interface User {
  id: number;
  username: string;
  balance: number;
}

export default function Wallet() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [depositAmount, setDepositAmount] = useState('');
  const [withdrawAmount, setWithdrawAmount] = useState('');
  const [showDepositModal, setShowDepositModal] = useState(false);
  const [showWithdrawModal, setShowWithdrawModal] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/auth/login');
      return;
    }

    // Загружаем данные пользователя
    Promise.all([
      fetch('/api/v1/users/profile', {
        headers: { 'Authorization': `Bearer ${token}` }
      }).then(res => res.json()),
      fetch('/api/v1/transactions', {
        headers: { 'Authorization': `Bearer ${token}` }
      }).then(res => res.json())
    ])
    .then(([userData, transactionData]) => {
      if (userData.success) {
        setUser(userData.data);
      }
      if (transactionData.success) {
        setTransactions(transactionData.data || []);
      }
    })
    .catch(error => {
      console.error('Error loading wallet data:', error);
    })
    .finally(() => setLoading(false));
  }, [router]);

  const getTransactionTypeText = (type: string) => {
    switch (type) {
      case 'deposit': return 'Пополнение';
      case 'withdraw': return 'Вывод';
      case 'purchase': return 'Покупка';
      case 'sale': return 'Продажа';
      default: return type;
    }
  };

  const getTransactionTypeColor = (type: string) => {
    switch (type) {
      case 'deposit': 
      case 'sale': 
        return 'text-green-600';
      case 'withdraw': 
      case 'purchase': 
        return 'text-red-600';
      default: 
        return 'text-gray-600';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'completed': return 'bg-green-100 text-green-800';
      case 'failed': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return 'Ожидание';
      case 'completed': return 'Завершено';
      case 'failed': return 'Ошибка';
      default: return status;
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
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const handleDeposit = async (e: React.FormEvent) => {
    e.preventDefault();
    // Здесь будет логика пополнения
    console.log('Deposit:', depositAmount);
    setShowDepositModal(false);
    setDepositAmount('');
  };

  const handleWithdraw = async (e: React.FormEvent) => {
    e.preventDefault();
    // Здесь будет логика вывода
    console.log('Withdraw:', withdrawAmount);
    setShowWithdrawModal(false);
    setWithdrawAmount('');
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header currentPath="/wallet" />
        <div className="max-w-6xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
          <div className="animate-pulse">
            <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
              <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
              <div className="h-12 bg-gray-200 rounded w-1/2"></div>
            </div>
            <div className="space-y-4">
              {[1, 2, 3].map(i => (
                <div key={i} className="bg-white rounded-lg shadow-sm p-6">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header currentPath="/wallet" />
      <div className="max-w-6xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Кошелек</h1>
          <p className="text-gray-600">Управляйте своими финансами</p>
        </div>

        {/* Баланс и действия */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-medium text-gray-900 mb-2">Текущий баланс</h2>
              <p className="text-4xl font-bold text-green-600">
                {user ? formatBalance(user.balance) : '0 ₽'}
              </p>
            </div>
            <div className="flex space-x-4">
              <button
                onClick={() => setShowDepositModal(true)}
                className="btn btn-primary"
              >
                <svg className="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Пополнить
              </button>
              <button
                onClick={() => setShowWithdrawModal(true)}
                className="btn btn-outline"
              >
                <svg className="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
                </svg>
                Вывести
              </button>
            </div>
          </div>
        </div>

        {/* История транзакций */}
        <div className="bg-white rounded-lg shadow-sm">
          <div className="p-6 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">История транзакций</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {transactions.length === 0 ? (
              <div className="p-12 text-center">
                <div className="mx-auto h-12 w-12 text-gray-400">
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                  </svg>
                </div>
                <h3 className="mt-2 text-sm font-medium text-gray-900">Нет транзакций</h3>
                <p className="mt-1 text-sm text-gray-500">
                  История ваших транзакций появится здесь
                </p>
              </div>
            ) : (
              transactions.map(transaction => (
                <div key={transaction.id} className="p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <div className={`flex-shrink-0 h-10 w-10 rounded-full flex items-center justify-center ${
                        transaction.type === 'deposit' || transaction.type === 'sale' 
                          ? 'bg-green-100' 
                          : 'bg-red-100'
                      }`}>
                        {transaction.type === 'deposit' || transaction.type === 'sale' ? (
                          <svg className="h-5 w-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                          </svg>
                        ) : (
                          <svg className="h-5 w-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
                          </svg>
                        )}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {getTransactionTypeText(transaction.type)}
                        </p>
                        <p className="text-sm text-gray-500">
                          {transaction.description}
                        </p>
                        <p className="text-xs text-gray-400">
                          {formatDate(transaction.created_at)}
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className={`text-lg font-semibold ${getTransactionTypeColor(transaction.type)}`}>
                        {transaction.type === 'deposit' || transaction.type === 'sale' ? '+' : '-'}
                        {formatBalance(transaction.amount)}
                      </p>
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(transaction.status)}`}>
                        {getStatusText(transaction.status)}
                      </span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Модальные окна */}
        {showDepositModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Пополнить баланс</h3>
              <form onSubmit={handleDeposit}>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Сумма пополнения
                  </label>
                  <input
                    type="number"
                    value={depositAmount}
                    onChange={(e) => setDepositAmount(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Введите сумму"
                    min="1"
                    required
                  />
                </div>
                <div className="flex justify-end space-x-3">
                  <button
                    type="button"
                    onClick={() => setShowDepositModal(false)}
                    className="btn btn-outline"
                  >
                    Отмена
                  </button>
                  <button type="submit" className="btn btn-primary">
                    Пополнить
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {showWithdrawModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Вывести средства</h3>
              <form onSubmit={handleWithdraw}>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Сумма вывода
                  </label>
                  <input
                    type="number"
                    value={withdrawAmount}
                    onChange={(e) => setWithdrawAmount(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Введите сумму"
                    min="1"
                    max={user?.balance || 0}
                    required
                  />
                  <p className="text-sm text-gray-500 mt-1">
                    Доступно: {user ? formatBalance(user.balance) : '0 ₽'}
                  </p>
                </div>
                <div className="flex justify-end space-x-3">
                  <button
                    type="button"
                    onClick={() => setShowWithdrawModal(false)}
                    className="btn btn-outline"
                  >
                    Отмена
                  </button>
                  <button type="submit" className="btn btn-primary">
                    Вывести
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
} 