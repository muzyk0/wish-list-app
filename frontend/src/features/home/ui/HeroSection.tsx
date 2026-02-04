"use client";

import Link from "next/link";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";

interface HeroSectionProps {
  onMobileRedirect: () => void;
}

export function HeroSection({ onMobileRedirect }: HeroSectionProps) {
  const { t } = useTranslation();

  return (
    <section className="relative min-h-[85vh] md:min-h-[90vh] flex items-center justify-center px-4 sm:px-6 overflow-hidden">
      {/* Background gradient orbs */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-80 h-80 md:w-[500px] md:h-[500px] rounded-full bg-gradient-to-br from-amber-200/30 to-orange-300/20 dark:from-amber-500/10 dark:to-orange-500/5 blur-3xl animate-pulse" />
        <div
          className="absolute -bottom-40 -left-40 w-80 h-80 md:w-[500px] md:h-[500px] rounded-full bg-gradient-to-tr from-rose-200/30 to-pink-300/20 dark:from-rose-500/10 dark:to-pink-500/5 blur-3xl animate-pulse"
          style={{ animationDelay: "1s" }}
        />
      </div>

      {/* Subtle grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.02] dark:opacity-[0.03]"
        style={{
          backgroundImage: `linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)`,
          backgroundSize: "60px 60px",
        }}
      />

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 px-4 py-2 mb-8 rounded-full bg-muted/50 dark:bg-muted/30 border border-border/50 backdrop-blur-sm animate-fade-in-up">
          <span className="size-2 rounded-full bg-gradient-to-r from-amber-500 to-orange-500 animate-pulse" />
          <span className="text-xs sm:text-sm font-medium text-muted-foreground tracking-wide">
            {t("hero.badge")}
          </span>
        </div>

        {/* Main heading */}
        <h1 className="text-4xl sm:text-5xl md:text-7xl lg:text-8xl font-bold tracking-tight mb-6 animate-fade-in-up animation-delay-100">
          <span className="block text-foreground">{t("hero.title.line1")}</span>
          <span className="block mt-2 bg-gradient-to-r from-amber-500 via-orange-500 to-rose-500 dark:from-amber-400 dark:via-orange-400 dark:to-rose-400 bg-clip-text text-transparent">
            {t("hero.title.line2")}
          </span>
        </h1>

        {/* Subtitle */}
        <p className="max-w-xl mx-auto text-base sm:text-lg md:text-xl text-muted-foreground mb-10 leading-relaxed animate-fade-in-up animation-delay-200">
          {t("hero.subtitle")}
        </p>

        {/* CTA buttons */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 animate-fade-in-up animation-delay-300">
          <Button
            onClick={onMobileRedirect}
            size="lg"
            className="w-full sm:w-auto min-w-[200px] h-12 sm:h-14 text-base font-semibold bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white shadow-lg shadow-orange-500/25 dark:shadow-orange-500/15 border-0 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl hover:shadow-orange-500/30"
          >
            <svg
              className="size-5 mr-2"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3"
              />
            </svg>
            {t("hero.openApp")}
          </Button>

          <Button
            variant="outline"
            size="lg"
            asChild
            className="w-full sm:w-auto min-w-[200px] h-12 sm:h-14 text-base font-semibold border-2 hover:bg-accent/50 transition-all duration-300"
          >
            <Link href="/auth/login">
              <svg
                className="size-5 mr-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z"
                />
              </svg>
              {t("hero.accountAccess")}
            </Link>
          </Button>
        </div>

        {/* Trust indicators */}
        <div className="mt-16 flex flex-wrap items-center justify-center gap-6 sm:gap-10 text-muted-foreground/60 animate-fade-in-up animation-delay-400">
          <div className="flex items-center gap-2">
            <svg className="size-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 22C6.477 22 2 17.523 2 12S6.477 2 12 2s10 4.477 10 10-4.477 10-10 10zm-1-11v6h2v-6h-2zm0-4v2h2V7h-2z" />
            </svg>
            <span className="text-sm font-medium">{t("hero.trust.free")}</span>
          </div>
          <div className="flex items-center gap-2">
            <svg className="size-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z" />
            </svg>
            <span className="text-sm font-medium">
              {t("hero.trust.secure")}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <svg className="size-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
            </svg>
            <span className="text-sm font-medium">
              {t("hero.trust.noDuplicates")}
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}
