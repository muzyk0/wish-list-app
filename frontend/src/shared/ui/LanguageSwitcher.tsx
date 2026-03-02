'use client';

import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/shared/ui/button';

export function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const [currentLanguage, setCurrentLanguage] = useState(i18n.language);

  useEffect(() => {
    setCurrentLanguage(i18n.language);

    const handleLanguageChange = () => {
      setCurrentLanguage(i18n.language);
    };

    // Подписываемся на событие изменения языка
    i18n.on('languageChanged', handleLanguageChange);

    // Отписываемся при размонтировании
    return () => {
      i18n.off('languageChanged', handleLanguageChange);
    };
  }, [i18n]);

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
  };

  return (
    <div className="flex gap-2">
      <Button
        variant={currentLanguage === 'en' ? 'default' : 'outline'}
        onClick={() => changeLanguage('en')}
        size="sm"
      >
        EN
      </Button>
      <Button
        variant={currentLanguage === 'ru' ? 'default' : 'outline'}
        onClick={() => changeLanguage('ru')}
        size="sm"
      >
        RU
      </Button>
    </div>
  );
}
