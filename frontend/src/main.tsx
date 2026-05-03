import * as Sentry from '@sentry/react';
import { ConfigProvider, theme } from 'antd';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { createIntl, IntlProvider } from 'react-intl';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import { cache, messages, useLocale } from './i18n/useLocale';
import './styles/index.css';

Sentry.init({
  dsn: 'https://b88f414e4f6248f0e601064aeff0b714@o81376.ingest.us.sentry.io/4510922576560128',
  integrations: [Sentry.browserTracingIntegration(), Sentry.replayIntegration()],
  tracesSampleRate: 1.0,
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,
  enableLogs: true,
  tracePropagationTargets: ['localhost', /^\/api\//],
});

const LocaleProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { locale } = useLocale();
  const intl = createIntl(
    {
      locale,
      messages: messages[locale],
    },
    cache
  );

  const antdLocale = locale === 'zh-CN' ? zhCN : enUS;

  return (
    <IntlProvider {...intl}>
      <ConfigProvider
        locale={antdLocale}
        theme={{
          algorithm: theme.defaultAlgorithm,
          token: {
            colorPrimary: '#1677ff',
          },
        }}
      >
        <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
          {children}
        </BrowserRouter>
      </ConfigProvider>
    </IntlProvider>
  );
};

const rootEl = document.getElementById('root');
if (!rootEl) throw new Error('Root element not found');
ReactDOM.createRoot(rootEl).render(
  <React.StrictMode>
    <LocaleProvider>
      <App />
    </LocaleProvider>
  </React.StrictMode>
);
