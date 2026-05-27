import WebApp from '@twa-dev/sdk';

export function initTelegram() {
  safeCall(() => WebApp.ready());
  safeCall(() => WebApp.expand());
  safeCall(() => WebApp.enableClosingConfirmation());

  document.documentElement.style.setProperty('--tg-bg', WebApp.themeParams.bg_color || '#f6f7fb');
  document.documentElement.style.setProperty('--tg-panel', WebApp.themeParams.secondary_bg_color || '#ffffff');
  document.documentElement.style.setProperty('--tg-text', WebApp.themeParams.text_color || '#161b22');
  document.documentElement.style.setProperty('--tg-hint', WebApp.themeParams.hint_color || '#6b7280');
  document.documentElement.style.setProperty('--tg-accent', WebApp.themeParams.button_color || '#2f80ed');
}

export function getInitData() {
  if (WebApp.initData) {
    return WebApp.initData;
  }

  const demoUser = encodeURIComponent(JSON.stringify({ id: 123456789, username: 'demo_user' }));
  return `user=${demoUser}&auth_date=1710000000`;
}

export function configureBackButton(visible: boolean, onClick: () => void) {
  safeCall(() => WebApp.BackButton.offClick(onClick));
  safeCall(() => WebApp.BackButton.onClick(onClick));
  if (visible) {
    safeCall(() => WebApp.BackButton.show());
  } else {
    safeCall(() => WebApp.BackButton.hide());
  }
}

export function configureMainButton(text: string, visible: boolean, onClick: () => void) {
  safeCall(() => WebApp.MainButton.offClick(onClick));
  if (!visible) {
    safeCall(() => WebApp.MainButton.hide());
    return;
  }

  safeCall(() => WebApp.MainButton.onClick(onClick));
  if (text.trim()) {
    safeCall(() => WebApp.MainButton.setText(text));
  }
  safeCall(() => WebApp.MainButton.show());
}

function safeCall(action: () => void) {
  try {
    action();
  } catch {
    // Telegram SDK methods can throw outside real Telegram clients.
  }
}
