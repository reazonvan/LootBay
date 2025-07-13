import { useState, useEffect } from 'react';
import Head from 'next/head';
import Link from 'next/link';
import Header from '../components/Header';

export default function HomePage() {
  const [apiStatus, setApiStatus] = useState('🔄 Проверка...');

  useEffect(() => {
    const checkAPI = async () => {
      try {
        const response = await fetch('/health');
        if (response.ok) {
          setApiStatus('🟢 Онлайн');
        } else {
          setApiStatus('🔴 Недоступно');
        }
      } catch (error) {
        setApiStatus('🔴 Ошибка');
      }
    };

    checkAPI();
  }, []);

  return (
    <>
      <Head>
        <title>LootBay - Безопасная торговля игровыми активами</title>
        <meta name="description" content="Современный маркетплейс для покупки и продажи игровых аккаунтов, валюты и предметов с системой escrow" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <div className="min-h-screen bg-gray-50">
        <Header currentPath="/" />

        {/* API Status */}
        <div className="bg-gray-100 py-2 px-4 text-center">
          <div className="text-sm">
            API Status: {apiStatus}
          </div>
        </div>

        {/* Hero Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto text-center">
            <h2 className="text-4xl md:text-6xl font-bold text-gray-900 mb-6">
              Безопасная торговля
              <span className="text-primary-600"> игровыми активами</span>
            </h2>
            
            <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
              LootBay - это современный маркетплейс для безопасной покупки и продажи 
              игровых аккаунтов, валюты и предметов с системой escrow и защитой сделок.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link href="/products" className="btn btn-primary btn-lg">
                Посмотреть товары
              </Link>
              <Link href="/auth/register" className="btn btn-outline btn-lg">
                Начать продавать
              </Link>
            </div>
          </div>
        </section>

        {/* Features */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h3 className="text-3xl font-bold text-gray-900 mb-4">
                Почему выбирают LootBay?
              </h3>
              <p className="text-xl text-gray-600">
                Мы предоставляем надежную платформу для безопасных сделок
              </p>
            </div>

            <div className="grid md:grid-cols-3 gap-8">
              <div className="text-center">
                <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                </div>
                <h4 className="text-xl font-semibold text-gray-900 mb-2">Безопасность</h4>
                <p className="text-gray-600">
                  Система escrow защищает ваши средства до успешного завершения сделки
                </p>
              </div>

              <div className="text-center">
                <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <h4 className="text-xl font-semibold text-gray-900 mb-2">Скорость</h4>
                <p className="text-gray-600">
                  Быстрые сделки и мгновенные уведомления о статусе ваших заказов
                </p>
              </div>

              <div className="text-center">
                <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192L5.636 18.364M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
                  </svg>
                </div>
                <h4 className="text-xl font-semibold text-gray-900 mb-2">Поддержка</h4>
                <p className="text-gray-600">
                  Круглосуточная поддержка и помощь в решении любых вопросов
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Footer */}
        <footer className="bg-secondary-900 text-white py-12">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="grid md:grid-cols-4 gap-8">
              <div>
                <h5 className="text-lg font-bold mb-4">LootBay</h5>
                <p className="text-gray-300">
                  Безопасный игровой маркетплейс для всех ваших потребностей в торговле.
                </p>
              </div>
              
              <div>
                <h6 className="font-semibold mb-4">Продукты</h6>
                <ul className="space-y-2 text-gray-300">
                  <li><Link href="/games" className="hover:text-white transition-colors">Игры</Link></li>
                  <li><Link href="/products" className="hover:text-white transition-colors">Товары</Link></li>
                  <li><Link href="/chat" className="hover:text-white transition-colors">Чат</Link></li>
                </ul>
              </div>
              
              <div>
                <h6 className="font-semibold mb-4">Поддержка</h6>
                <ul className="space-y-2 text-gray-300">
                  <li><Link href="/help" className="hover:text-white transition-colors">Помощь</Link></li>
                  <li><Link href="/contact" className="hover:text-white transition-colors">Контакты</Link></li>
                  <li><Link href="/faq" className="hover:text-white transition-colors">FAQ</Link></li>
                </ul>
              </div>
              
              <div>
                <h6 className="font-semibold mb-4">Компания</h6>
                <ul className="space-y-2 text-gray-300">
                  <li><Link href="/about" className="hover:text-white transition-colors">О нас</Link></li>
                  <li><Link href="/terms" className="hover:text-white transition-colors">Условия</Link></li>
                  <li><Link href="/privacy" className="hover:text-white transition-colors">Конфиденциальность</Link></li>
                </ul>
              </div>
            </div>
            
            <div className="border-t border-gray-700 mt-8 pt-8 text-center text-gray-300">
              <p>&copy; 2024 LootBay. Все права защищены.</p>
            </div>
          </div>
        </footer>
      </div>
    </>
  );
} 