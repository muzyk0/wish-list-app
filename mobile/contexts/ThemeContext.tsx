import type React from "react";
import { createContext, useContext, useEffect, useState } from "react";
import { Appearance } from "react-native";
import { darkTheme, lightTheme, type ThemeType } from "@/theme";

interface ThemeContextType {
  theme: ThemeType;
  isDark: boolean;
  toggleTheme: () => void;
  setTheme: (dark: boolean) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const useThemeContext = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useThemeContext must be used within a ThemeProvider");
  }
  return context;
};

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [isDark, setIsDark] = useState(Appearance.getColorScheme() === "dark");
  const [theme, setTheme] = useState<ThemeType>(
    isDark ? darkTheme : lightTheme,
  );

  useEffect(() => {
    // Listen to system theme changes
    const subscription = Appearance.addChangeListener(({ colorScheme }) => {
      setIsDark(colorScheme === "dark");
    });

    return () => {
      subscription.remove();
    };
  }, []);

  useEffect(() => {
    // Update theme when isDark changes
    setTheme(isDark ? darkTheme : lightTheme);
  }, [isDark]);

  const toggleTheme = () => {
    setIsDark((prev) => !prev);
  };

  const setThemeManually = (dark: boolean) => {
    setIsDark(dark);
  };

  const value = {
    theme,
    isDark,
    toggleTheme,
    setTheme: setThemeManually,
  };

  return (
    <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
  );
};
