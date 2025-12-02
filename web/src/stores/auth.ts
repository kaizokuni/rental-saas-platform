import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import client from '../api/client';

interface User {
  id: string;
  email: string;
  role: string;
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'));
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'));

  const isAuthenticated = computed(() => !!token.value);

  function setAuth(newToken: string, newUser: User) {
    token.value = newToken;
    user.value = newUser;
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
  }

  function logout() {
    token.value = null;
    user.value = null;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }

  async function login(email: string, password: string) {
    const response = await client.post('/api/auth/login', { email, password });
    setAuth(response.data.token, response.data.user);
  }

  return {
    token,
    user,
    isAuthenticated,
    setAuth,
    login,
    logout,
  };
});
