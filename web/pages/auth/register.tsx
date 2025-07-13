import { useState, useEffect } from 'react';
import Head from 'next/head';
import Link from 'next/link';
import { useRouter } from 'next/router';
import dynamic from 'next/dynamic';
import { motion, AnimatePresence } from 'framer-motion';

const InputMask = dynamic(() => import('react-input-mask'), {
  ssr: false,
});

// Тип политики пароля, получаемой с сервера
interface PasswordPolicy {
  min_length: number;
  required_categories: number;
  categories: string[];
}

// Проверяем пароль на соответствие политике (минимальная длина и n категорий)
const isPasswordValid = (password: string, policy: PasswordPolicy | null): boolean => {
  if (!policy) return true; // если не успели получить — не блокируем, проверит сервер

  if (password.length < policy.min_length) return false;

  const checks: Record<string, boolean> = {
    lower: /[a-z]/.test(password),
    upper: /[A-Z]/.test(password),
    digit: /\d/.test(password),
    special: /[!@#$%^&*(),.?":{}|<>\-_=+\[\]\\/`~;]/.test(password),
  };

  // Подсчитываем, сколько категорий пройдено
  const passed = policy.categories.reduce((acc, cat) => acc + (checks[cat] ? 1 : 0), 0);
  return passed >= policy.required_categories;
};

export default function RegisterPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    email: '',
    phone: '',
    username: '',
    password: '',
    confirmPassword: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [policy, setPolicy] = useState<PasswordPolicy | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // При монтировании запрашиваем политику паролей
  useEffect(() => {
    fetch('/api/v1/config/password-policy')
      .then((r) => r.json())
      .then((data) => setPolicy(data.password_policy))
      .catch(() => null); // в случае ошибки просто игнорируем, сервер проверит
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  // Функция для преобразования телефона в формат E.164
  const formatPhoneToE164 = (phone: string): string => {
    // Убираем все символы кроме цифр и +
    const cleaned = phone.replace(/[^\d+]/g, '');
    // Если номер начинается с 8, заменяем на +7
    if (cleaned.startsWith('8')) {
      return '+7' + cleaned.substring(1);
    }
    // Если номер начинается с 7, добавляем +
    if (cleaned.startsWith('7')) {
      return '+' + cleaned;
    }
    // Если номер начинается с +, оставляем как есть
    if (cleaned.startsWith('+')) {
      return cleaned;
    }
    // По умолчанию добавляем +7
    return '+7' + cleaned;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // Клиентская валидация пароля
    if (!isPasswordValid(formData.password, policy)) {
      setError('Пароль не соответствует требованиям');
      return;
    }

    if (formData.password !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const response = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: formData.email,
          phone: formatPhoneToE164(formData.phone),
          username: formData.username,
          password: formData.password,
        }),
      });

      const data = await response.json().catch(() => null);
      if (response.ok) {
        router.push('/auth/login');
      } else {
        // Улучшенная обработка ошибок
        let errorMessage = 'Ошибка регистрации';
        if (data?.error?.code === 'USER_EXISTS') {
          errorMessage = 'Пользователь с таким email уже существует';
        } else if (data?.error?.code === 'USERNAME_EXISTS') {
          errorMessage = 'Пользователь с таким именем уже существует';
        } else if (data?.error?.code === 'PHONE_EXISTS') {
          errorMessage = 'Пользователь с таким телефоном уже существует';
        } else if (data?.error?.message) {
          errorMessage = data.error.message;
        }
        setError(errorMessage);
      }
    } catch (err) {
      setError('Ошибка соединения с сервером');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Head>
        <title>Регистрация - LootBay</title>
        <meta name="description" content="Создайте аккаунт LootBay" />
      </Head>

      <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="sm:mx-auto sm:w-full sm:max-w-md"
        >
          <div className="text-center">
            <Link href="/" className="text-3xl font-bold text-primary-600">
              LootBay
            </Link>
          </div>
          <h2 className="mt-6 text-center text-3xl font-bold text-gray-900">
            Создайте новый аккаунт
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Или{' '}
            <Link href="/auth/login" className="font-medium text-primary-600 hover:text-primary-500">
              войдите в аккаунт
            </Link>
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="mt-8 sm:mx-auto sm:w-full sm:max-w-md"
        >
          <div className="card">
            <div className="card-body">
              <form onSubmit={handleSubmit} className="space-y-6">
                <AnimatePresence>
                  {error && (
                    <motion.div
                      initial={{ opacity: 0, y: -10, height: 0 }}
                      animate={{ opacity: 1, y: 0, height: 'auto' }}
                      exit={{ opacity: 0, y: -10, height: 0 }}
                      transition={{ duration: 0.3 }}
                      className="bg-red-50 border border-red-200 rounded-lg p-4"
                    >
                      <p className="text-sm text-red-800">{error}</p>
                    </motion.div>
                  )}
                </AnimatePresence>

                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                    Email
                  </label>
                  <div className="mt-1">
                    <input
                      id="email"
                      name="email"
                      type="email"
                      required
                      value={formData.email}
                      onChange={handleChange}
                      className="input transition-shadow duration-200"
                      placeholder="Введите email"
                    />
                  </div>
                </div>

                <div>
                  <label htmlFor="phone" className="block text-sm font-medium text-gray-700">
                    Телефон
                  </label>
                  <div className="mt-1">
                    {/* Используем InputMask для красивого форматирования номера телефона */}
                    <InputMask
                      mask="+7 (999) 999-99-99"
                      maskChar={null}
                      id="phone"
                      name="phone"
                      type="tel"
                      required
                      value={formData.phone}
                      onChange={handleChange}
                      className="input transition-shadow duration-200"
                      placeholder="+7 (999) 123-45-67" // красивый пример
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Формат: +7 (999) 123-45-67
                    </p>
                  </div>
                </div>

                <div>
                  <label htmlFor="username" className="block text-sm font-medium text-gray-700">
                    Имя пользователя
                  </label>
                  <div className="mt-1">
                    <input
                      id="username"
                      name="username"
                      type="text"
                      required
                      value={formData.username}
                      onChange={handleChange}
                      className="input transition-shadow duration-200"
                      placeholder="Введите имя пользователя"
                    />
                  </div>
                </div>

                <div>
                  <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                    Пароль
                  </label>
                  <div className="mt-1 relative">
                    <input
                      id="password"
                      name="password"
                      type={showPassword ? 'text' : 'password'}
                      required
                      value={formData.password}
                      onChange={handleChange}
                      className="input pr-12 transition-shadow duration-200"
                      placeholder="Введите пароль"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500"
                    >
                      <AnimatePresence initial={false} mode="wait">
                        <motion.div
                          key={showPassword ? 'eye' : 'eye-slash'}
                          initial={{ y: 10, opacity: 0 }}
                          animate={{ y: 0, opacity: 1 }}
                          exit={{ y: -10, opacity: 0 }}
                          transition={{ duration: 0.2 }}
                        >
                          {showPassword ? (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                            </svg>
                          ) : (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.542-7 1.274-4.057 5.064-7 9.542-7 1.356 0 2.65.31 3.838.868M6.375 9.375a9.954 9.954 0 011.625-2.025M18.625 14.625a9.954 9.954 0 01-1.625 2.025" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15a3 3 0 100-6 3 3 0 000 6z" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21L3 3" />
                            </svg>
                          )}
                        </motion.div>
                      </AnimatePresence>
                    </button>
                  </div>
                  {/* Подсказка с требованиями политики */}
                  {policy && (
                    <p className="mt-1 text-xs text-gray-500">
                      Минимум {policy.min_length} символов, любые {policy.required_categories} из&nbsp;
                      {policy.categories.join(', ')}
                    </p>
                  )}
                </div>

                <div>
                  <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700">
                    Подтвердите пароль
                  </label>
                  <div className="mt-1 relative">
                    <input
                      id="confirmPassword"
                      name="confirmPassword"
                      type={showConfirmPassword ? 'text' : 'password'}
                      required
                      value={formData.confirmPassword}
                      onChange={handleChange}
                      className="input pr-12 transition-shadow duration-200"
                      placeholder="Повторите пароль"
                    />
                    <button
                      type="button"
                      onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                      className="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500"
                    >
                      <AnimatePresence initial={false} mode="wait">
                        <motion.div
                          key={showConfirmPassword ? 'eye' : 'eye-slash'}
                          initial={{ y: 10, opacity: 0 }}
                          animate={{ y: 0, opacity: 1 }}
                          exit={{ y: -10, opacity: 0 }}
                          transition={{ duration: 0.2 }}
                        >
                          {showConfirmPassword ? (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                            </svg>
                          ) : (
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.542-7 1.274-4.057 5.064-7 9.542-7 1.356 0 2.65.31 3.838.868M6.375 9.375a9.954 9.954 0 011.625-2.025M18.625 14.625a9.954 9.954 0 01-1.625 2.025" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15a3 3 0 100-6 3 3 0 000 6z" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21L3 3" />
                            </svg>
                          )}
                        </motion.div>
                      </AnimatePresence>
                    </button>
                  </div>
                </div>

                <div>
                  <motion.button
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                    type="submit"
                    disabled={loading}
                    className="btn btn-primary w-full btn-md"
                  >
                    {loading ? 'Создание...' : 'Создать аккаунт'}
                  </motion.button>
                </div>
              </form>
            </div>
          </div>
        </motion.div>
      </div>
    </>
  );
} 