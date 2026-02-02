'use client';

import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
  };

  return (
    <div className="flex gap-2">
      <Button
        variant={i18n.language === 'en' ? 'default' : 'outline'}
        onClick={() => changeLanguage('en')}
        size="sm"
      >
        EN
      </Button>
      <Button
        variant={i18n.language === 'ru' ? 'default' : 'outline'}
        onClick={() => changeLanguage('ru')}
        size="sm"
      >
        RU
      </Button>
    </div>
  );
}