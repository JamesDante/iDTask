import { API_BASE_URL } from "@/config";

export const apiClient = {
  post: async (path: string, data: any) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`API POST failed: ${response.status}`);
    }

    return response.json();
  },
};
