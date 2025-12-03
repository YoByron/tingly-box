// API service layer for communicating with the backend

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

// Get auth token from localStorage
const getAuthToken = (): string | null => {
  return localStorage.getItem('auth_token');
};

async function fetchUIAPI(url: string, options: RequestInit = {}): Promise<any> {
  try {
    const fullUrl = url.startsWith('/api/') ? url : `/api${url}`;
    const token = getAuthToken();

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>,
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(fullUrl, {
      headers,
      ...options,
    });

    // Handle 401 Unauthorized - token is invalid or expired
    if (response.status === 401) {
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
      return { success: false, error: 'Authentication required' };
    }

    return await response.json();
  } catch (error) {
    console.error('UI API Error:', error);
    return { success: false, error: (error as Error).message };
  }
}

async function fetchServerAPI(url: string, options: RequestInit = {}): Promise<any> {
  try {
    const fullUrl = url.startsWith('/api/') ? API_BASE_URL + url : url;
    const token = getAuthToken();

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>,
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(fullUrl, {
      headers,
      ...options,
    });

    // Handle 401 Unauthorized - token is invalid or expired
    if (response.status === 401) {
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
      return { success: false, error: 'Authentication required' };
    }

    return await response.json();
  } catch (error) {
    console.error('Server API Error:', error);
    return { success: false, error: (error as Error).message };
  }
}

export const api = {
  // Status endpoints
  getStatus: () => fetchUIAPI('/status'),
  getProviders: () => fetchUIAPI('/providers'),
  getProviderModels: () => fetchUIAPI('/provider-models'),
  getProviderModelsByName: (name: string) => fetchUIAPI(`/provider-models/${name}`),
  getDefaults: () => fetchUIAPI('/defaults'),
  setDefaults: (data: any) => fetchUIAPI('/defaults', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  getHistory: (limit?: number) => fetchUIAPI(`/history${limit ? `?limit=${limit}` : ''}`),

  // Provider management
  addProvider: (data: any) => fetchUIAPI('/providers', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  getProvider: (name: string) => fetchUIAPI(`/providers/${name}`),
  updateProvider: (name: string, data: any) => fetchUIAPI(`/providers/${name}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
  deleteProvider: (name: string) => fetchUIAPI(`/providers/${name}`, {
    method: 'DELETE',
  }),
  toggleProvider: (name: string) => fetchUIAPI(`/providers/${name}/toggle`, {
    method: 'POST',
  }),

  // Server control
  startServer: (port: number) => fetchServerAPI('/api/server/start', {
    method: 'POST',
    body: JSON.stringify({ port }),
  }),
  stopServer: () => fetchServerAPI('/api/server/stop', { method: 'POST' }),
  restartServer: (port: number) => fetchServerAPI('/api/server/restart', {
    method: 'POST',
    body: JSON.stringify({ port }),
  }),
  generateToken: (clientId: string) => fetchServerAPI(`/api/token?client_id=${encodeURIComponent(clientId)}`),
};

export default api;
