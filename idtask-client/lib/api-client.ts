import { API_BASE_URL } from "@/config";

export const apiClient = {
  
  post: async (path: string, data: any) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: "POST",
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`API POST failed: ${response.status}`);
    }

    return response.json();
  },

  get: async <T = any>(path: string): Promise<T> => {
    const response = await fetch(`${API_BASE_URL}${path}`);
    if (!response.ok) {
      throw new Error(`API GET failed: ${response.status}`);
    }
    return response.json();
  },
};
