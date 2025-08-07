import { QueryClient, type QueryKey } from "@tanstack/react-query";

// Default fetcher function for GET requests
const defaultQueryFn = async ({ queryKey }: { queryKey: QueryKey }) => {
  const url = Array.isArray(queryKey) ? queryKey[0] : queryKey;
  const response = await fetch(url as string);
  if (!response.ok) {
    throw new Error(`Network response was not ok: ${response.statusText}`);
  }
  return response.json();
};

// Create and export the query client
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      queryFn: defaultQueryFn,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

// API request helper for mutations (POST, PATCH, DELETE)
export const apiRequest = async (
  url: string,
  options: RequestInit = {}
): Promise<any> => {
  const response = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.statusText}`);
  }

  return response.json();
};