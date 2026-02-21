import React from 'react'
import ReactDOM from 'react-dom/client'
import * as Sentry from '@sentry/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, theme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import { IntlProvider } from 'react-intl'
import App from './App'
import { useLocale, cache, messages } from './i18n/useLocale'
import { createIntl } from 'react-intl'
import './styles/index.css'

Sentry.init({
  dsn: 'https://b88f414e4f6248f0e601064aeff0b714@o81376.ingest.us.sentry.io/4510922576560128',
  integrations: [
    Sentry.browserTracingIntegration(),
    Sentry.replayIntegration(),
  ],
  tracesSampleRate: 1.0,
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,
  enableLogs: true,
  tracePropagationTargets: ['localhost', /^\/api\//],
})

const LocaleProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { locale } = useLocale()
  const intl = createIntl(
    {
      locale,
      messages: messages[locale],
    },
    cache
  )

  const antdLocale = locale === 'zh-CN' ? zhCN : enUS

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
        <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>{children}</BrowserRouter>
      </ConfigProvider>
    </IntlProvider>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <LocaleProvider>
      <App />
    </LocaleProvider>
  </React.StrictMode>,
)
