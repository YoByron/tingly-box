import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface AuthContextType {
  token: string | null;
  isAuthenticated: boolean;
  login: (token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [token, setToken] = useState<string | null>(() => {
    return localStorage.getItem('auth_token');
  });

  const isAuthenticated = !!token;

  const login = (newToken: string) => {
    setToken(newToken);
    localStorage.setItem('auth_token', newToken);
  };

  const logout = () => {
    setToken(null);
    localStorage.removeItem('auth_token');
  };

  useEffect(() => {
    const storedToken = localStorage.getItem('auth_token');
    if (storedToken !== token) {
      setToken(storedToken);
    }
  }, []);

  return (
    <AuthContext.Provider value={{ token, isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};